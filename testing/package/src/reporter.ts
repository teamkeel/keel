import { TestCaseResult } from "output";
import fetch from 'node-fetch'

export interface ReporterOptions {
  host: string;
  port: number;
}

export default class Reporter {
  private readonly opts : ReporterOptions;

  constructor(opts: ReporterOptions) {
    this.opts = opts;
  }

  report = async (results: TestCaseResult[]) : Promise<boolean> => {
    const response = await this.doRequest(results)

    return response.ok
  }

  private async doRequest(results: TestCaseResult[]) {
    const { port, host } = this.opts;

    return await fetch(`http://${host}:${port}/report`, { method: 'POST', body: JSON.stringify(results) })
  }
}
