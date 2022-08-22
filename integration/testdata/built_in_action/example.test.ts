import { test, expect, Actions } from '@teamkeel/testing'

// createPost is a built in operation

test('it passes', async () => {
  const result = await Actions.createPost({ title: 'foo' })
  expect.equal(result.title, 'foo')
})

test('it fails', async () => {
  const result = await Actions.createPost({ title: 'foo' })

  expect.equal(result.title, 'bar')
})

