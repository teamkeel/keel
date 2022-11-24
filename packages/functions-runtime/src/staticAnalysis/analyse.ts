import { parseFiles } from "@structured-types/api";

export interface AnalyseOptions {
  path: string;
}

export default async (opts: AnalyseOptions) => {
  return parseFiles([opts.path], { collectParameters: true });
};
