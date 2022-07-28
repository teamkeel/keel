import { program } from 'commander'
import util from 'util'
import path from 'path';

const glob = util.promisify(require('glob'))

import analyse from './analyse'

program
  .version('0.0.1')
  .description('Statically analyses a custom function directory')
  .argument('<string>', 'Path to directory')
  .option('-d, --debug', 'Debug mode')
  .action(async (dir, opts) => {
    const pattern = path.join(dir, 'functions', '*.ts')
    const tsFiles: string[] = await glob(pattern)

    let json = {}

    for await (const result of tsFiles.map(
      async (p) => ({ [p]: await analyse({ path: p }) }))
    ) {
      json = { ...json, ...result }
    }

    if (opts.debug) {
      console.log(util.inspect(json, { showHidden: false, colors: true, depth: 5 }))
    } else {
      console.log('%j', json)
    }
  })

program.parse()
