import { resetDatabase, flows, models } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";
import { File } from "@teamkeel/sdk";

beforeEach(resetDatabase);
test("flows - file inputs flow", async () => {
  let flow = await flows.fileInput.start({});
  expect(flow).toEqual({
    id: expect.any(String),
    traceId: expect.any(String),
    status: "AWAITING_INPUT",
    name: "FileInput",
    startedBy: null,
    input: {},
    error: null,
    data: null,
    steps: [
      {
        id: expect.any(String),
        name: "file input page",
        runId: expect.any(String),
        stage: null,
        status: "PENDING",
        type: "UI",
        value: null,
        error: null,
        startTime: expect.any(Date),
        endTime: null,
        createdAt: expect.any(Date),
        updatedAt: expect.any(Date),
        ui: {
          __type: "ui.page",
          content: [
            {
              __type: "ui.input.file",
              label: "Avatar",
              disabled: false,
              helpText: "A nice photo of yourself",
              name: "avatar",
              optional: true,
            },
            {
              __type: "ui.input.file",
              disabled: false,
              label: "Passport",
              name: "passport",
              optional: false,
            },
          ],
          hasValidationErrors: false,
        },
      },
    ],
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
    config: {
      title: "File input",
    },
  });

  const avatarCallbackResponse = await flows.fileInput.callback(
    flow.id,
    flow.steps[0].id,
    "avatar",
    "getPresignedUploadURL",
    {}
  );
  expect(avatarCallbackResponse).toEqual({
    key: expect.any(String),
    url: expect.any(String),
  });

  const passportCallbackResponse = await flows.fileInput.callback(
    flow.id,
    flow.steps[0].id,
    "passport",
    "getPresignedUploadURL",
    { key: "existing-file-key" }
  );
  expect(passportCallbackResponse).toEqual({
    key: "existing-file-key",
    url: expect.any(String),
  });

  const imageData = `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`;

  // upload the avatar to the presigned upload Url
  const response = await fetch(avatarCallbackResponse.url, {
    method: "PUT",
    headers: {
      "Content-Type": "image/png",
    },
    body: Buffer.from(imageData, "base64"),
  });

  expect(response.ok).toBe(true);

  // put the page step values
  flow = await flows.fileInput.putStepValues(flow.id, flow.steps[0].id, {
    avatar: {
      key: avatarCallbackResponse.key,
      filename: "my-avatar.png",
      contentType: "image/png",
      size: Buffer.from(imageData, "base64").length,
    },
  });

  // complete the flow
  const completedFlow = await flows.fileInput.untilFinished(flow.id);

  expect(completedFlow.status).toBe("COMPLETED");
  expect(completedFlow.steps[0].status).toBe("COMPLETED");

  // first ctx.step to create user and return its id and avatarURL
  expect(completedFlow.steps[1].name).toBe("create user");
  expect(completedFlow.steps[1].status).toBe("COMPLETED");
  expect(completedFlow.steps[1].value).toEqual({
    userId: expect.any(String),
    avatarUrl: expect.any(String),
  });
  const userId = completedFlow.steps[1].value.userId;

  // second ctx.step to retrieve user from db, get the image and return a presigned url for the image
  expect(completedFlow.steps[2].name).toBe("get image");
  expect(completedFlow.steps[2].status).toBe("COMPLETED");

  // thw whole flow to return the id of the user and a presigned url for the uploaded avatar
  expect(completedFlow.data.id).toBe(userId);
  expect(completedFlow.data.avatarUrl).toContain(
    `/aws/files/${avatarCallbackResponse.key}?`
  );

  // retrieve model from db
  const user = await models.user.findOne({ id: userId });
  expect(user).toEqual({
    id: expect.any(String),
    avatar: expect.any(File),
    passport: null,
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
  });

  //assert that the contents read from the File retrieved via the ModelAPI are the same as the ones we uploaded
  const contents1 = await user!.avatar?.read();
  const base64Contents = contents1.toString("base64");
  expect(base64Contents).toContain(imageData);
});
