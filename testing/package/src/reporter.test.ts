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

        const testResults = TestResult.pass({
          name: "a passing test",
          fn: () => {},
          filePath: "/foo/bar"
        }) as TestResult;

        await reporter.reportResult([testResults]);

        expect(fetchMock.mock.calls[0][1]).toEqual({
          method: "POST",
          body: '[{"name":"a passing test","filePath":"/foo/bar","status":"pass"}]',
        });
        expect(fetchMock.mock.calls[0][0]).toEqual(
          "http://localhost:123/report"
        );
      });
    });

    describe("with a failing test", () => {
      it("sends a request to the correct url", async () => {
        fetchMock.mockResponseOnce(
          JSON.stringify({
            object: {
              title: "text",
            },
          })
        );

        const reporter = new Reporter({ port: 123, host: "localhost" });

        const testResults = TestResult.fail({
          name: "a failing test",
          fn: () => {},
          filePath: "/foo/bar"
        }, 1, 2) as TestResult;

        await reporter.reportResult([testResults]);

        expect(fetchMock.mock.calls[0][1]).toEqual({
          method: "POST",
          body: '[{"name":"a failing test","filePath":"/foo/bar","status":"fail","expected":2,"actual":1}]',
        });
        expect(fetchMock.mock.calls[0][0]).toEqual(
          "http://localhost:123/report"
        );
      });
    });

    describe("with a test raising an exception", () => {
      it("sends a request to the correct url", async () => {
        fetchMock.mockResponseOnce(
          JSON.stringify({
            object: {
              title: "text",
            },
          })
        );

        const reporter = new Reporter({ port: 123, host: "localhost" });

        const err = new Error('oops')
        const testResults = TestResult.exception({
          name: "a test that errors",
          fn: () => {},
          filePath: "/foo/bar"
        }, err) as TestResult;

        await reporter.reportResult([testResults]);

        const expectedJsonPartial = JSON.parse('{"name":"a test that errors","filePath":"/foo/bar","status":"exception"}')
        const body = fetchMock.mock.calls![0]![1]!.body
        const actualJson = JSON.parse(body!.toString())

        // we cannot assert that the error details match
        // as they are non deterministic so just check for a 
        // partial match for the rest of the keys
        expect(actualJson[0]).toEqual(
          expect.objectContaining(expectedJsonPartial)
        )
        expect(fetchMock.mock.calls[0][0]).toEqual(
          "http://localhost:123/report"
        );
      });
    });
  });
});
