import { CreateAccountInput, API } from "@teamkeel/sdk";

export default async (inputs: CreateAccountInput, api: API) => {
  // Build your universe

  api.models.Account.create({ name: "my account" });
};
