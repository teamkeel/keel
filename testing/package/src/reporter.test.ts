import { TestResult } from "./output";
import Reporter from "./reporter";

describe("reporter", () => {
  describe("report", () => {
    beforeEach(() => {
      fetchMock.resetMocks();
    });

    describe("with a passing test", () => {
      it("sends a request to the correct url", async () => {
        fetchMock.mockResponseOnce(
          JSON.stringify({
            object: {
              title: "text",
            },
          })
        );

        const reporter = new Reporter({ port: 123, host: "localhost" });

        const testResults = [TestResult.pass("a passing test")] as TestResult[];

        await reporter.report(testResults);

        expect(fetchMock.mock.calls[0][1]).toEqual({
          method: "POST",
          body: '[{"testName":"a passing test","status":"pass"}]',
        });
        expect(fetchMock.mock.calls[0][0]).toEqual(
          "http://localhost:123/report"
        );
      });
    });
  });
});
