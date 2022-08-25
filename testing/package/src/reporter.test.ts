import { TestResultData } from "output";
import Reporter from "./reporter";

describe("reporter", () => {
  beforeEach(() => {
    fetchMock.resetMocks();
  });

  it("sends a request to the correct url", async () => {
    fetchMock.mockResponseOnce(
      JSON.stringify({
        result: {
          title: "text",
        },
      })
    );

    const reporter = new Reporter({ port: 123, host: "localhost" });

    const testResults = [
      {
        testName: "a passing test",
        status: "pass",
      },
    ] as TestResultData[];
    const result = await reporter.report(testResults);

    expect(fetchMock.mock.calls[0][1]).toEqual({
      method: "POST",
      body: '[{"testName":"a passing test","status":"pass"}]',
    });
    expect(fetchMock.mock.calls[0][0]).toEqual("http://localhost:123/report");
  });
});
