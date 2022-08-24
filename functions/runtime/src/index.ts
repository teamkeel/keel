import { createServer, IncomingMessage, ServerResponse } from "http";
import url from "url";

import { Config } from "./types";

const startRuntimeServer = ({ functions, api }: Config) => {
  const listener = async (req: IncomingMessage, res: ServerResponse) => {
    if (req.method === "POST") {
      const parts = url.parse(req.url!);
      const { pathname } = parts;

      const normalisedPathname = pathname!.replace(/\//, "");

      const buffers = [];

      for await (const chunk of req) {
        buffers.push(chunk);
      }

      const data = Buffer.concat(buffers).toString();

      const json = JSON.parse(data);

      const { call } = functions[normalisedPathname];

      try {
        const result = await call(json, api);

        // We want to do an explicit check for null or undefined here
        // as some actions can return a boolean (delete actions)
        if (result === undefined || result === null) {
          throw new Error(`No value returned from ${normalisedPathname}`);
        }

        res.write(JSON.stringify({ result }));
      } catch (e) {
        if (e instanceof Error) {
          const { message } = e;

          res.write(JSON.stringify({ error: message }));
        } else {
          res.write(JSON.stringify({ error: "An unknown error occurred" }));
        }

        res.statusCode = 500;
      }
    } else {
      res.statusCode = 400;

      res.write(
        JSON.stringify({
          message: "Only POST requests are permitted",
        })
      );
    }

    res.end();
  };

  const server = createServer(listener);

  const port = (process.env.PORT && parseInt(process.env.PORT, 10)) || 3001;

  server.listen(port);
};

export default startRuntimeServer;
