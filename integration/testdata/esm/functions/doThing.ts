import { DoThing } from "@teamkeel/sdk";
import fetch from "node-fetch";

export default DoThing(async (inputs, api, ctx) => {
  const a = await fetch("google.com");

  const person = await api.models.person.findOne(inputs);
  return person;
});
