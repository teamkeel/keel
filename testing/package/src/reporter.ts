import { TestResult } from "./output";
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

  clearDatabase = async ({ filePath, testCase }: ClearOptions): Promise<boolean> => {
    const res = await fetch(`${this.buildHostUri()}/reset`, {
      method: "POST",
      body: JSON.stringify({
        filePath,
        testCase,
      })
    });

    return res.ok;
  };

  report = async (results: TestResult[]): Promise<boolean> => {
    const response = await this.testResultsRequest(results);
    return response.ok;
  };

  private async testResultsRequest(results: TestResult[]) {
    return await fetch(`${this.buildHostUri()}/report`, {
      method: "POST",
      // JSON.stringify will call TestResult.toJSON for each result in the array
      body: JSON.stringify(results),
    });
  }

  private buildHostUri = () => {
    const { port, host } = this.opts;

    return `http://${host}:${port}`;
  };
}
