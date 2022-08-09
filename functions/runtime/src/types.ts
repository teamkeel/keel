export interface CustomFunction {
  call: any
}

// Config represents the configuration values
// to be passed to the Custom Code runtime server
export interface Config {
  functions: Record<string, CustomFunction>
  api: unknown
}
