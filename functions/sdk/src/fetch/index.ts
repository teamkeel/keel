import fetch, { RequestInit } from "node-fetch";

type RequestOpts = Omit<RequestInit, "referrer">;

export default async (uri: string, opts: RequestOpts) => {
  // todo: more orchestration
  return fetch(uri, opts);
};
