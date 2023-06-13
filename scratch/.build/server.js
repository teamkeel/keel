import { handleRequest, tracing } from '@teamkeel/functions-runtime';
import { createContextAPI, permissionFns } from '@teamkeel/sdk';
import { createServer } from "http";
import function_createPerson from "../functions/createPerson.ts";
const functions = {
    createPerson: function_createPerson,
}
const actionTypes = {
    createPerson: "OPERATION_TYPE_CREATE",
}

const listener = async (req, res) => {
	const u = new URL(req.url, "http://" + req.headers.host);
	if (req.method === "GET" && u.pathname === "/_health") {
		res.statusCode = 200;
		res.end();
		return;
	}

	if (req.method === "POST") {
		const buffers = [];
		for await (const chunk of req) {
			buffers.push(chunk);
		}
		const data = Buffer.concat(buffers).toString();
		const json = JSON.parse(data);

		const rpcResponse = await handleRequest(json, {
			functions,
			createContextAPI,
			actionTypes,
			permissionFns,
		});

		res.statusCode = 200;
		res.setHeader('Content-Type', 'application/json');
		res.write(JSON.stringify(rpcResponse));
		res.end();
		return;
	}

	res.statusCode = 400;
	res.end();
};

tracing.init();

const server = createServer(listener);
const port = (process.env.PORT && parseInt(process.env.PORT, 10)) || 3001;
server.listen(port);
