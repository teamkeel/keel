import { TestResult } from "./output";
import { TestCase } from "./types";
import fetch from "node-fetch";

export interface ReporterOptions {
  host: string;
  port: number;
}

interface ClearOptions {
  filePath: string;
  testCase: string;
}

export default class Reporter {
  private readonly opts: ReporterOptions;

  constructor(opts: ReporterOptions) {
    this.opts = opts;
  }

  // Clear the database prior to running each test casde
  clearDatabase = async ({
    filePath,
    testCase,
  }: ClearOptions): Promise<boolean> => {
    const res = await fetch(`${this.buildHostUri()}/reset`, {
      method: "POST",
      body: JSON.stringify({
        filePath,
        testCase,
      }),
    });

    return res.ok;
  };

  // At the beginning of the test run, report all of the known
  // tests back to the go process, so it can output them in the TUI
  collectTests = async (tests: TestCase[]): Promise<boolean> => {
    const res = await fetch(`${this.buildHostUri()}/collect`, {
      method: "POST",
      body: JSON.stringify(tests),
    });

    return res.ok;
  };

  // Report the test result (pass | fail | exception) when the test function
  // has been evaluated in the try/catch block
  reportResult = async (results: TestResult[]): Promise<boolean> => {
    const response = await fetch(`${this.buildHostUri()}/report`, {
      method: "POST",
      // JSON.stringify will call TestResult.toJSON for each result in the array
      body: JSON.stringify(results.map((r) => r.toJSON())),
    });
    return response.ok;
  };

  private buildHostUri = () => {
    const { port, host } = this.opts;

    return `http://${host}:${port}`;
  };
}
