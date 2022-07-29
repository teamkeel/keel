import { createServer, IncomingMessage, ServerResponse } from 'http'
import url from 'url'

import { Config } from "./types"
import buildApi from './api'

const startRuntimeServer = (config: Config) => {
  const { functions, models } = config

  const listener = async (req: IncomingMessage, res: ServerResponse) => {
    if (req.method === 'POST') {
      const parts = url.parse(req.url!)
      const { pathname } = parts

      const normalisedPathname = pathname!.replace(/\//, "")

      const buffers = [];

      for await (const chunk of req) {
        buffers.push(chunk);
      }

      const data = Buffer.concat(buffers).toString();

      const json = JSON.parse(data)

      const { call } = functions[normalisedPathname]

      const result = await call(json, buildApi(models))

      console.log(JSON.stringify(result))
      
      res.write(JSON.stringify(result))
      
      res.end()
    }
  }
  
  const server = createServer(listener)

  const port = process.env.PORT && parseInt(process.env.PORT, 10) || 3001

  server.listen(port, 'localhost', 2, () => {
    console.log('server listening')
  })
}

export default startRuntimeServer
