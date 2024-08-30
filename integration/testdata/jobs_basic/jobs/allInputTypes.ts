import { AllInputTypes, InlineFile, Status } from "@teamkeel/sdk";
import { IdentityUniqueConditions } from "../.build/sdk/index";

export default AllInputTypes(async (ctx, inputs) => {
  if (inputs.text != "text") {
    throw new Error("text not set correctly");
  }
  if (inputs.num != 10) {
    throw new Error("num not set correctly");
  }
  if (inputs.boolean != true) {
    throw new Error("bool not set correctly");
  }
  if (inputs.date == null) {
    throw new Error("date not set correctly");
  }
  if (inputs.timestamp == null) {
    throw new Error("timestamp not set correctly");
  }
  if (inputs.id != "123") {
    throw new Error("id not set correctly");
  }
  if (inputs.enum != Status.GoldPost) {
    throw new Error("enum not set correctly");
  }
  if (JSON.stringify(inputs.array) != JSON.stringify(["one", "two"])) {
    throw new Error("array not set correctly");
  }
  // if (inputs.image.filename != "my-avatar.png") {
  //   throw new Error("image filename not set correctly");
  // }
  // if (inputs.image.size != 2024) {
  //   throw new Error("image not set correctly");
  // }
  // const imgData = await inputs.image.read();
  // if (imgData.byteLength != 2024) {
  //   throw new Error("image not set correctly");
  // }
});
