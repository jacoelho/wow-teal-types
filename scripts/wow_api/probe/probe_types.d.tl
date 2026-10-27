local record Types
   record Family
      name: string
      canonical: string
      methods: {string}
   end
   record Plan
      version: string
      expectedBuild: string
      sourceFingerprint: string
      declarationFingerprint: string
      expectedMethods: integer
      expectedInterface: integer
      expectedProjectID: integer
      fixtureSHA256: string
      limits: {string: integer}
      families: {Family}
      hierarchy: {HierarchyRequirement}
      candidateEvidence: CandidateEvidence
      nativeCandidateEvidence: NativeCandidateEvidence
      representatives: {RepresentativeRequirement}
      bindingDiagnosticCases: {string}
      bindingTextSampleCases: {string}
      bindingExperimentCases: {string}
   end
   record NativeCandidateEvidence
      capture: string
      sha256: string
      build: string
      role: string
   end
   record RepresentativeRequirement
      canonical: string
      resourceKey: string
      constructor: string
   end
   record CandidateEvidence
      repository: string
      commit: string
      build: string
      path: string
      git_blob_sha1: string
      sha256: string
      role: string
      methods: {string}
   end
   record HierarchyRequirement
      canonical: string
      parent: string
      methods: {string}
   end
   record Observation
      caseID: string
      state: string
      value: string | boolean
   end
   record Invocation
      caseID: string
      state: string
      invoked: boolean
      returnCount: integer
      returns: {string}
      check: string
      behaviorVerified: boolean
   end
   record BindingDiagnostic
      caseID: string
      state: string
   end
   record BindingTextSample
      caseID: string
      state: string
      hex: string
      byteLength: integer
      firstNul: integer
   end
   record FamilyResult
      family: string
      factory: string
      methods: {Observation}
      queries: {Observation}
   end
   record Report
      schema: integer
      probeVersion: string
      expectedBuild: string
      sourceFingerprint: string
      declarationFingerprint: string
      clientVersion: string
      clientBuild: string
      clientDate: string
      clientInterface: number
      projectID: number
      context: string
      outcome: string
      expectedMethods: integer
      limits: {string: integer}
      observedMethods: integer
      presentMethods: integer
      capabilityCaptureComplete: boolean
      verifiesSignatures: boolean
      verifiesSecrecy: boolean
      cleanup: {Observation}
      families: {FamilyResult}
      mode: string
      invocations: {Invocation}
      invokedMethods: integer
      verifiedMethods: integer
      invocationCaptureComplete: boolean
      fixtureSHA256: string
      hierarchy: {Observation}
      bindingDiagnostics: {BindingDiagnostic}
      bindingTextSamples: {BindingTextSample}
      bindingExperimentChecks: {BindingDiagnostic}
   end
   record InvocationHost
      field: function(any, string): (any, string)
      resolve: function(string, ? string): (any, string)
      create: function(any, any...): (any, string)
      construct: function(string, {string: any}, {string: string}): (any, string)
      uiObject: function(string): (any, string)
      observeHierarchy: function(any, string, ? string): {Observation}
   end
   record Namespace
      Plan: Plan
      RunInvocations: function(InvocationHost, Report)
   end
end
return Types
