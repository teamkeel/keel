import { test, expect, Actions } from '@teamkeel/testing'

test('it passes', async () => {
  const result = await Actions.createPerson({ title: 'foo' })
  expect.equal(result.title, 'foo')
})

test('it fails', async () => {
  const result = await Actions.createPerson({ title: 'bar' })
  expect.equal(result.title, 'foo')
})
