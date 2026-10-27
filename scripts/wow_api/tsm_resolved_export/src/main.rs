use bincode::Options;
use serde_json::{Value, json};
use sha2::{Digest, Sha256};
use std::{collections::BTreeMap, env, fs, io::Read};
use wowlua_analysis::pre_globals::{
    BLOB_MAGIC, BLOB_VERSION, PreResolvedGlobals, PrecomputedStubs,
};
use wowlua_core::types::{Expr, FunctionIndex, SymbolIdentifier, TableIndex, ValueType};

const SUPPORTED_BLOB_VERSION: u32 = 44;
const MAX_BLOB_SIZE: u64 = 64 * 1024 * 1024;
const SOURCE_LOCK: &str = include_str!("../../tsm_sources.json");
const CARGO_MANIFEST: &str = include_str!("../Cargo.toml");
const CARGO_LOCK: &str = include_str!("../Cargo.lock");

struct ReaderIdentity {
    commit: String,
    blob_version: u32,
    blob_size: u64,
    blob_sha256: String,
}

fn hex_id(value: &str, length: usize) -> bool {
    value.len() == length
        && value
            .bytes()
            .all(|byte| byte.is_ascii_digit() || (b'a'..=b'f').contains(&byte))
}

fn reader_identity(
    lock: &str,
    manifest: &str,
    cargo_lock: &str,
) -> Result<ReaderIdentity, Box<dyn std::error::Error>> {
    let source: Value = serde_json::from_str(lock)?;
    let commit = source["commit"]
        .as_str()
        .ok_or("missing resolved TSM commit")?;
    if source["repository"] != "TradeSkillMaster/wowlua-ls" || !hex_id(commit, 40) {
        return Err("invalid resolved TSM source identity".into());
    }
    let version = source["blob_version"]
        .as_u64()
        .ok_or("missing resolved TSM blob version")?;
    // Updating source identity cannot opt an existing decoder into another schema.
    if version != u64::from(SUPPORTED_BLOB_VERSION) || BLOB_VERSION != SUPPORTED_BLOB_VERSION {
        return Err(format!("unsupported resolved TSM reader schema: {version}").into());
    }
    let files = source["files"]
        .as_array()
        .ok_or("missing resolved TSM source files")?;
    let blobs: Vec<_> = files
        .iter()
        .filter(|row| row["name"] == "precomputed.bin.zst")
        .collect();
    if blobs.len() != 1 {
        return Err("resolved TSM source lock requires one resolved blob".into());
    }
    let size = blobs[0]["size"]
        .as_u64()
        .ok_or("missing resolved TSM blob size")?;
    let sha = blobs[0]["sha256"]
        .as_str()
        .ok_or("missing resolved TSM blob hash")?;
    if !(8..=MAX_BLOB_SIZE).contains(&size) || !hex_id(sha, 64) {
        return Err("invalid resolved TSM blob identity".into());
    }
    let manifest: toml::Value = toml::from_str(manifest)?;
    let cargo_lock: toml::Value = toml::from_str(cargo_lock)?;
    let repo = "https://github.com/TradeSkillMaster/wowlua-ls";
    let dependencies = manifest
        .get("dependencies")
        .and_then(toml::Value::as_table)
        .ok_or("missing Cargo dependencies")?;
    for name in ["wowlua_analysis", "wowlua_core"] {
        let dependency = dependencies
            .get(name)
            .ok_or_else(|| format!("missing Cargo dependency: {name}"))?;
        if dependency.get("git").and_then(toml::Value::as_str) != Some(repo)
            || dependency.get("rev").and_then(toml::Value::as_str) != Some(commit)
            || dependency.get("path").is_some()
            || dependency.get("branch").is_some()
            || dependency.get("tag").is_some()
        {
            return Err(format!("resolved TSM Cargo revision mismatch: {name}").into());
        }
    }
    let packages = cargo_lock
        .get("package")
        .and_then(toml::Value::as_array)
        .ok_or("missing Cargo package locks")?;
    for name in ["wowlua_analysis", "wowlua_core", "wowlua_syntax"] {
        let selected: Vec<_> = packages
            .iter()
            .filter(|row| row.get("name").and_then(toml::Value::as_str) == Some(name))
            .collect();
        let expected = format!("git+{repo}?rev={commit}#{commit}");
        if selected.len() != 1
            || selected[0].get("source").and_then(toml::Value::as_str) != Some(expected.as_str())
        {
            return Err(format!("resolved TSM Cargo.lock revision mismatch: {name}").into());
        }
    }
    Ok(ReaderIdentity {
        commit: commit.to_owned(),
        blob_version: version as u32,
        blob_size: size,
        blob_sha256: sha.to_owned(),
    })
}

fn table_type(value: &ValueType) -> Result<TableIndex, String> {
    match value {
        ValueType::Table(Some(index)) => Ok(*index),
        _ => Err(format!("binding is not one concrete table: {value:?}")),
    }
}

fn function_type(value: &ValueType) -> Result<FunctionIndex, String> {
    match value {
        ValueType::Function(Some(index)) => Ok(*index),
        _ => Err(format!("binding is not one concrete function: {value:?}")),
    }
}

fn selected(pg: &PreResolvedGlobals, name: &str) -> Result<FunctionIndex, String> {
    let parts: Vec<_> = name.split('.').collect();
    let mut table = None;
    if let Some(index) = pg
        .scope0_symbols
        .get(&SymbolIdentifier::Name(parts[0].to_owned()))
    {
        let symbol = pg.sym(*index);
        if symbol.versions.len() != 1 {
            return Err(format!("ambiguous serialized symbol versions: {name}"));
        }
        let value = symbol.versions[0]
            .resolved_type
            .as_ref()
            .ok_or_else(|| format!("unresolved global: {name}"))?;
        if parts.len() == 1 {
            return function_type(value);
        }
        table = Some(table_type(value)?);
    } else if parts.len() > 1 {
        table = pg.classes.get(parts[0]).copied();
    }
    let mut table = table.ok_or_else(|| format!("missing resolved binding: {name}"))?;
    for (position, part) in parts.iter().enumerate().skip(1) {
        let field = pg
            .table(table)
            .fields
            .get(*part)
            .ok_or_else(|| format!("missing resolved field: {name}"))?;
        let value = match pg.try_expr(field.expr) {
            Some(Expr::FunctionDef(index)) if position + 1 == parts.len() => return Ok(*index),
            Some(Expr::TableConstructor(index)) => {
                table = *index;
                continue;
            }
            Some(Expr::Literal(value)) => value,
            Some(Expr::SymbolRef(index, version)) => pg
                .sym(*index)
                .versions
                .get(*version)
                .and_then(|value| value.resolved_type.as_ref())
                .ok_or_else(|| format!("unresolved field reference: {name}"))?,
            other => {
                return Err(format!(
                    "unsupported resolved field expression for {name}: {other:?}"
                ));
            }
        };
        if position + 1 == parts.len() {
            return function_type(value);
        }
        table = table_type(value)?;
    }
    Err(format!("binding did not reach a function: {name}"))
}

fn source_hashes() -> BTreeMap<&'static str, String> {
    [
        ("Cargo.toml", include_str!("../Cargo.toml")),
        ("Cargo.lock", include_str!("../Cargo.lock")),
        ("src/main.rs", include_str!("main.rs")),
        ("tsm_sources.json", SOURCE_LOCK),
    ]
    .into_iter()
    .map(|(name, data)| (name, format!("{:x}", Sha256::digest(data.as_bytes()))))
    .collect()
}

fn run() -> Result<(), Box<dyn std::error::Error>> {
    let identity = reader_identity(SOURCE_LOCK, CARGO_MANIFEST, CARGO_LOCK)?;
    let path = env::args()
        .nth(1)
        .ok_or("usage: tsm-resolved-export BLOB REQUESTS")?;
    let mut raw = Vec::new();
    fs::File::open(path)?
        .take(identity.blob_size + 1)
        .read_to_end(&mut raw)?;
    if raw.len() as u64 != identity.blob_size
        || format!("{:x}", Sha256::digest(&raw)) != identity.blob_sha256
    {
        return Err("resolved TSM blob integrity failure".into());
    }
    if raw.len() < 8
        || u32::from_le_bytes(raw[..4].try_into()?) != BLOB_MAGIC
        || u32::from_le_bytes(raw[4..8].try_into()?) != BLOB_VERSION
        || BLOB_VERSION != identity.blob_version
    {
        return Err("incompatible resolved TSM blob header".into());
    }
    let mut data = Vec::new();
    zstd::stream::read::Decoder::new(&raw[8..])?
        .take(256 * 1024 * 1024 + 1)
        .read_to_end(&mut data)?;
    if data.len() > 256 * 1024 * 1024 {
        return Err("resolved TSM blob exceeds decode limit".into());
    }
    let stubs: PrecomputedStubs = bincode::DefaultOptions::new()
        .with_fixint_encoding()
        .with_limit(256 * 1024 * 1024)
        .reject_trailing_bytes()
        .deserialize(&data)?;
    let request_path = env::args().nth(2).ok_or("missing requests JSON")?;
    let mut request_data = Vec::new();
    fs::File::open(request_path)?
        .take(1024 * 1024 + 1)
        .read_to_end(&mut request_data)?;
    if request_data.len() > 1024 * 1024 {
        return Err("requests exceed 1 MiB".into());
    }
    let requests: Vec<Value> = serde_json::from_slice(&request_data)?;
    if requests.len() > 7000 {
        return Err("invalid request count".into());
    }
    let pg = &stubs.pre_globals;
    let mut rows = BTreeMap::new();
    for request in requests {
        let symbol = request["symbol"].as_str().ok_or("request lacks symbol")?;
        let tsm_symbol = request["tsm_symbol"]
            .as_str()
            .ok_or("request lacks tsm_symbol")?;
        if rows.contains_key(symbol) {
            return Err(format!("duplicate request: {symbol}").into());
        }
        let index = selected(pg, tsm_symbol)?;
        let location = pg
            .function_locations
            .get(&index)
            .ok_or_else(|| format!("selected function has no serialized source: {tsm_symbol}"))?;
        let function = pg.func(index);
        let args: Vec<_> = function.args.iter().enumerate().map(|(position, index)| {
            let symbol = pg.sym(*index);
            json!({"name": symbol.id, "version": symbol.versions.get(function.param_version(position))})
        }).collect();
        rows.insert(symbol.to_owned(), json!({"symbol": symbol, "tsm_symbol": tsm_symbol, "location": location, "function": function, "argument_symbols": args}));
    }
    serde_json::to_writer(
        std::io::stdout().lock(),
        &json!({"schema_version": 1, "tsm_commit": identity.commit, "blob_version": identity.blob_version,
                "blob_size": identity.blob_size, "blob_sha256": identity.blob_sha256,
                "exporter_sources": source_hashes(), "bindings": rows}),
    )?;
    Ok(())
}

fn main() {
    if let Err(error) = run() {
        eprintln!("{error}");
        std::process::exit(1);
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn release_fixture() -> (Value, toml::Value, toml::Value) {
        let lock = json!({"repository": "TradeSkillMaster/wowlua-ls",
            "commit": "0123456789abcdef0123456789abcdef01234567", "blob_version": 44,
            "files": [{"name": "precomputed.bin.zst", "size": 32,
                "sha256": "abababababababababababababababababababababababababababababababab"}]});
        let manifest = toml::from_str(r#"
[dependencies]
wowlua_analysis = { git = "https://github.com/TradeSkillMaster/wowlua-ls", rev = "0123456789abcdef0123456789abcdef01234567" }
wowlua_core = { git = "https://github.com/TradeSkillMaster/wowlua-ls", rev = "0123456789abcdef0123456789abcdef01234567" }
"#).unwrap();
        let cargo_lock = toml::from_str(r#"
[[package]]
name = "wowlua_analysis"
source = "git+https://github.com/TradeSkillMaster/wowlua-ls?rev=0123456789abcdef0123456789abcdef01234567#0123456789abcdef0123456789abcdef01234567"
[[package]]
name = "wowlua_core"
source = "git+https://github.com/TradeSkillMaster/wowlua-ls?rev=0123456789abcdef0123456789abcdef01234567#0123456789abcdef0123456789abcdef01234567"
[[package]]
name = "wowlua_syntax"
source = "git+https://github.com/TradeSkillMaster/wowlua-ls?rev=0123456789abcdef0123456789abcdef01234567#0123456789abcdef0123456789abcdef01234567"
"#).unwrap();
        (lock, manifest, cargo_lock)
    }

    fn check_fixture(
        lock: &Value,
        manifest: &toml::Value,
        cargo_lock: &toml::Value,
    ) -> Result<ReaderIdentity, Box<dyn std::error::Error>> {
        reader_identity(
            &lock.to_string(),
            &toml::to_string(manifest)?,
            &toml::to_string(cargo_lock)?,
        )
    }

    #[test]
    fn supported_release_identity_is_data_not_a_second_hardcoded_pin() {
        let (lock, manifest, cargo_lock) = release_fixture();
        let identity = check_fixture(&lock, &manifest, &cargo_lock).unwrap();
        assert_eq!(identity.commit, "0123456789abcdef0123456789abcdef01234567");
        assert_eq!(identity.blob_version, 44);
        assert_eq!(identity.blob_size, 32);
        assert_eq!(
            identity.blob_sha256,
            "abababababababababababababababababababababababababababababababab"
        );
    }

    #[test]
    fn changing_a_source_lock_does_not_opt_into_an_unknown_schema() {
        let (mut lock, manifest, cargo_lock) = release_fixture();
        lock["blob_version"] = json!(45);
        let error = check_fixture(&lock, &manifest, &cargo_lock).err().unwrap();
        assert!(
            error
                .to_string()
                .contains("unsupported resolved TSM reader schema")
        );
    }

    #[test]
    fn source_identity_requires_one_bounded_hash_identified_blob() {
        let (lock, manifest, cargo_lock) = release_fixture();
        for files in [
            json!([]),
            json!([lock["files"][0].clone(), lock["files"][0].clone()]),
        ] {
            let mut changed = lock.clone();
            changed["files"] = files;
            assert!(check_fixture(&changed, &manifest, &cargo_lock).is_err());
        }
        for (key, value) in [
            ("size", json!(0)),
            ("size", json!(true)),
            ("size", json!(64 * 1024 * 1024 + 1)),
            ("sha256", json!("invalid")),
            (
                "sha256",
                json!("ABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABABAB"),
            ),
        ] {
            let mut changed = lock.clone();
            changed["files"][0][key] = value;
            assert!(check_fixture(&changed, &manifest, &cargo_lock).is_err());
        }
    }

    #[test]
    fn every_compiled_dependency_must_match_the_canonical_cohort() {
        let (lock, manifest, cargo_lock) = release_fixture();
        for name in ["wowlua_analysis", "wowlua_core"] {
            for (key, value) in [
                ("rev", "ffffffffffffffffffffffffffffffffffffffff"),
                ("git", "https://example.invalid/other"),
            ] {
                let mut changed = manifest.clone();
                changed["dependencies"][name][key] = toml::Value::String(value.to_owned());
                let error = check_fixture(&lock, &changed, &cargo_lock).err().unwrap();
                assert!(error.to_string().contains("Cargo revision mismatch"));
            }
        }
        for position in 0..3 {
            let mut changed = cargo_lock.clone();
            changed["package"][position]["source"] = toml::Value::String(
                "git+https://example.invalid/other#ffffffffffffffffffffffffffffffffffffffff"
                    .to_owned(),
            );
            let error = check_fixture(&lock, &manifest, &changed).err().unwrap();
            assert!(error.to_string().contains("Cargo.lock revision mismatch"));
        }
    }

    #[test]
    fn missing_binding_never_uses_a_location_candidate() {
        let pg = PreResolvedGlobals::empty();
        assert!(
            selected(&pg, "Missing")
                .unwrap_err()
                .contains("missing resolved binding")
        );
        assert!(
            selected(&pg, "Missing.Method")
                .unwrap_err()
                .contains("missing resolved binding")
        );
    }

    #[test]
    fn only_concrete_function_bindings_are_accepted() {
        for value in [
            ValueType::Any,
            ValueType::Function(None),
            ValueType::Union(vec![ValueType::Function(None), ValueType::Nil]),
        ] {
            assert!(function_type(&value).is_err());
        }
    }

    #[test]
    fn only_concrete_namespace_bindings_are_accepted() {
        for value in [
            ValueType::Any,
            ValueType::Table(None),
            ValueType::Union(vec![ValueType::Table(None), ValueType::Nil]),
        ] {
            assert!(table_type(&value).is_err());
        }
    }
}
