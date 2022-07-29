export type MapSchemaTypes = {
  string: string;
  integer: number;
  // others?
}

export type MapSchema<T extends Record<string, keyof MapSchemaTypes>> = {
  -readonly [K in keyof T]: MapSchemaTypes[T[K]]
}

export function asSchema<T extends Record<string, keyof MapSchemaTypes>>(t: T): T {
  return t;
}

export interface CustomFunction {
  // We don't know the exact types of the inputs and api at this point in the toolchain
  call: (inputs: unknown, api: unknown) => Promise<unknown>
}

export interface Model {
  name: string
  definition: Record<string, keyof MapSchemaTypes>
}

// Config represents the configuration values
// to be passed to the Custom Code runtime server
export interface Config {
  functions: Record<string, CustomFunction>
  models: Model[]
}

export interface BaseAPI {
  models: Record<string, unknown>
}

export interface ModelApi<T> {
  create: (inputs: Partial<Omit<T, "id" | "createdAt" | "updatedAt">>) => Promise<T>
  delete: (id: string) => Promise<boolean>
  find: (p: Partial<T>) => Promise<T>
  update: (id: string, inputs: Partial<Omit<T, "id" | "createdAt" | "updatedAt">>) => Promise<T>
  findMany: (p: Partial<T>) => Promise<T[]>
}
