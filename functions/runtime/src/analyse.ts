import { parseFiles } from '@structured-types/api';

export interface AnalyseOptions {
  path: string
}

export default (opts: AnalyseOptions) => {
  return parseFiles([opts.path])
}
