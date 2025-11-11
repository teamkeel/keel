import { resetDatabase, models, flows } from "@teamkeel/testing";
import { beforeEach, expect, test } from "vitest";

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

  // client would now upload to the `passportCallbackResponse.url` and `avatarCallbackResponse.url` the
  // files, and then submit the page

  flow = await flows.fileInput.putStepValues(flow.id, flow.steps[0].id, {
    avatar: {
      key: avatarCallbackResponse.key,
      filename: "my-avatar.png",
      contentType: "image/png",
    },
  });

  const completedFlow = await flows.fileInput.untilFinished(flow.id);

  expect(completedFlow.status).toBe("COMPLETED");
  expect(completedFlow.steps[0].status).toBe("COMPLETED");
  expect(completedFlow.steps[1].status).toBe("COMPLETED");
  expect(completedFlow.steps[1].value).toBeTruthy();

  // retrieve model from db
  const user = await models.user.findOne({ id: completedFlow.steps[1].value });
  expect(user).toEqual({
    id: expect.any(String),
    avatar: {
      key: avatarCallbackResponse.key,
      filename: "my-avatar.png",
      contentType: "image/png",
      size: expect.any(Number),
    },
    passport: null,
    createdAt: expect.any(Date),
    updatedAt: expect.any(Date),
  });
});
