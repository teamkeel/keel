import { parseFiles } from '@structured-types/api';

export interface AnalyseOptions {
  path: string
}

export default async (opts: AnalyseOptions) => {
  // parseFiles may take an array of paths, but it actually only works
  // for a single path
  // Passing in multiple file paths with default exports within each doesnt work
  return parseFiles([opts.path], { collectParameters: true })
}
