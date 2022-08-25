import ActionExecutor from './actionExecutor'

describe('execute an action', () => {
  beforeEach(() => {
    fetchMock.resetMocks()
  })

  it('succeeds', async () => {
    fetchMock.mockResponseOnce(JSON.stringify({
      result: {
        title: 'text'
      }
    }))
    const executor = new ActionExecutor({ parentPort: 123, protocol: 'http', host: 'localhost' });
  
    const result = await executor.execute({ actionName: 'createPost', payload: { title: 'a post' } });

    expect(fetchMock.mock.calls[0][0]).toEqual('http://localhost:123/action')
    expect(result).toEqual({
      result: {
        title: 'text'
      }
    })
  })
})