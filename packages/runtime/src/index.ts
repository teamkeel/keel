import { createServer, IncomingMessage, ServerResponse } from "http";
import url from "url";

import { Config } from "./types";

let blahahahahahhahah = "";








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
        // Call the custom function
        // Every custom function has an enforced return type
        const result = await call(json, api);

        // Handle if no value is returned by the custom function
        if (!result) {
          res.write(
            JSON.stringify({
              errors: [
                {
                  message: `No value returned from ${normalisedPathname}`,
                },
              ],
            })
          );

          return;
        }

        // successful result will be:
        // {
        //   object: XXX,
        //   errors: [...]
        // }
        res.write(JSON.stringify(result));
      } catch (e) {
        // Catch unhandled errors
        console.error(e);
        res.statusCode = 500;

        if (e instanceof Error) {
          const { message } = e;

          res.write(
            JSON.stringify({
              errors: [
                {
                  message,
                },
              ],
            })
          );
        } else {
          res.write(
            JSON.stringify({
              errors: [
                {
                  message: "An unknown error occurred",
                },
              ],
            })
          );
        }
      }
    } else {
      res.statusCode = 400;

      res.write(
        JSON.stringify({
          errors: [
            {
              message: "Only POST requests are permitted",
            },
          ],
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
