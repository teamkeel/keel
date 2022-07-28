export type Timestamp = string

export interface Post {
  title: string
  id: string
  createdAt: Date
  updatedAt: Date
}

export interface Identity {
  username: string
  id: string
  createdAt: Date
  updatedAt: Date
}

export interface PostInputs {
  title: string
}
export interface API {
  models: {
    Post: PostApi

  }
}

export interface PostApi {
  create: (inputs: PostInputs) => Promise<Post>
  delete: (id: string) => Promise<boolean>
  find: (p: Partial<Post>) => Promise<Post>
  update: (id: string, inputs: PostInputs) => Promise<Post>
  findMany: (p: Partial<Post>) => Promise<Post[]>
}
interface CustomFunction {
  call: any
  contextModel: string
}

// Config represents the configuration values
// to be passed to the Custom Code runtime server
interface Config {
  functions: Record<string, CustomFunction>
}

import { createServer, IncomingMessage, ServerResponse } from 'http'

declare global {
  namespace NodeJS {
    interface ProcessEnv {
      PORT?: string;
    }
  }
}

import url from 'url'

const startServer = (config: Config) => {
  const listener = async (req: IncomingMessage, res: ServerResponse) => {
    if (req.method === 'POST') {
      const parts = url.parse(req.url)
      const { pathname } = parts
      const normalisedPathname = pathname.replace(/\//, "")

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

  server.listen(parseInt(process.env.PORT, 10), 'localhost', 2, () => {
    console.log('server listening')
  })
}
import createPost from '../functions/createPost'


startServer({
  functions: { createPost: { call: createPost, contextModel: 'Post' },  },
})
