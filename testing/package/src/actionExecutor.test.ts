import ActionExecutor from "./actionExecutor";

describe("execute an action", () => {
  beforeEach(() => {
    fetchMock.resetMocks();
  });

  describe('happy path', () => {
    it("succeeds", async () => {
      fetchMock.mockResponseOnce(
        JSON.stringify({
          object: {
            title: "text",
          },
        })
      );
      const executor = new ActionExecutor({
        parentPort: 123,
        protocol: "http",
        host: "localhost",
      });
  
      const result = await executor.execute({
        actionName: "createPost",
        payload: { title: "a post" },
      });
  
      expect(fetchMock.mock.calls[0][0]).toEqual("http://localhost:123/action");
      expect(result).toEqual({
        object: {
          title: "text",
        },
      });
    });
  });

  describe('when the action has errors in the response', () => {
    it('throws the errors', async () => {
      fetchMock.mockResponseOnce(
        JSON.stringify({
          object: {
            title: "text",
          },
          errors: [
            {
              message: 'first error'
            },
            {
              message: 'second error'
            }
          ]
        })
      );
      const executor = new ActionExecutor({
        parentPort: 123,
        protocol: "http",
        host: "localhost",
      });
  

      await expect(async() => await executor.execute({
        actionName: "createPost",
        payload: { title: "a post" },
      })).rejects.toThrowError('first error, second error')
    });
  })
});
