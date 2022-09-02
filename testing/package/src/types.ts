export type TestFunc = () => void | Promise<void>;
export type TestName = string;

export interface Test {
  testName: TestName;
  fn: TestFunc;
}

export interface RunnerOpts {
  parentPort: number;
  host: string;
  debug?: boolean;
}

export type ScalarTypes = "string" | "boolean" | "date" | "number";

export type ModelDefinition<T> = Record<keyof T, ScalarTypes>;
