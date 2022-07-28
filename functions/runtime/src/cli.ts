import { program } from 'commander'
import util from 'util'
import analyse from './analyse'

program
  .version('0.0.1')
  .description('Statically analyses a custom function directory')
  .argument('<string>', 'Path to file')
  .option('-d, --debug', 'Debug mode')
  .action(async (path, opts) => {
    const result = await analyse({ path }) 

    if (opts.debug) {
      console.log(util.inspect(result, { showHidden: false, colors: true, depth: null }))
    } else {
      console.log('%j', result)
    }
  })

program.parse()
