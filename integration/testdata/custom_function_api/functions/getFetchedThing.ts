import { GetFetchedThing, FetchedThing } from "@teamkeel/sdk";

export default GetFetchedThing(async (inputs, api) => {
  const response = await api.fetch("http://example.com/movies.json'");
  const bodyStr = await response.text();

  var ft: FetchedThing = {
    fetchedBody: bodyStr,
    id: "unused id",
    createdAt: new Date(),
    updatedAt: new Date(),
  };

  return new Promise((resolve) => {
    resolve(ft);
  });
});
