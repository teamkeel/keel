import { createServer, IncomingMessage, ServerResponse } from 'http'
import url from 'url'

import { Config } from "../../types"

const startRuntimeServer = (config: Config) => {
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

      const { call, contextModel } = config.functions[normalisedPathname]

      // todo: place all models here
      const api = {
        models: {
          [contextModel]: {
            create: async () => ({
              id: 123,
              title: json.title
            })
          }
        }
      }
      const result = await call(json, api)
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
