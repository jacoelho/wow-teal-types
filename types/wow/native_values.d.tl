-- Static restriction markers; Secret does not assert a runtime userdata carrier.
local record Values
   record Secret<T>
      is userdata
   end
   type MaybeSecret<T> = T | Secret<T>
   type Data = number | string | boolean | {any: any} | function(any...): any...
   type Atom = number | string | boolean | function(any...): any...
   type SecretData = MaybeSecret<Atom> | Secret<SecretTable>
   type SecretTable = {SecretData: SecretData}
end
return Values
