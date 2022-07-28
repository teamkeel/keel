import { program } from 'commander'
import analyse from './analyse'

program
  .version('0.0.1')
  .description('Statically analyses a custom function file')
  .argument('<string>', 'Path to file')
  .action((path) => {
    console.log('%j', analyse({ path }))
  })

program.parse()
