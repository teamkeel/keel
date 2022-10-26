import { test, expect, actions, Post, logger } from '@teamkeel/testing'
import { LogLevel } from '@teamkeel/sdk'

test('List Where filters - using equal operator (string) - filters correctly', async () => {
  await actions.createPost({ title: 'Fred' })
  await actions.createPost({ title: 'NotFred' })

  const { collection: response } = await actions.listPostsEqualString({ 
      whereArg: "Fred" 
    })

  expect(response.length).toEqual(1)
  expect(response[0].title).toEqual('Fred')
})


test('List Where filters - using not equal operator (string) - filters correctly', async () => {
  await actions.createPost({ title: 'Fred' })
  await actions.createPost({ title: 'NotFred' })

  const { collection: response } = await actions.listPostsNotEqualString({ 
      whereArg: "Fred"
    })

  expect(response.length).toEqual(1)
  expect(response[0].title).toEqual('NotFred')
})

test('List Where filters - using equal operator on date - filters correctly', async () => {
  await actions.createPost({ aDate: new Date(2020, 1, 21) })
  await actions.createPost({ aDate: new Date(2020, 1, 22) })
  await actions.createPost({ aDate: new Date(2020, 1, 23) })

  const { collection: response } = await actions.listPostsEqualDate({ 
      whereArg: new Date(2020, 1, 21)
  })

  expect(response.length).toEqual(1)
})

test('List Where filters - using not equal operator on date - filters correctly', async () => {
  await actions.createPost({ aDate: new Date(2020, 1, 21) })
  await actions.createPost({ aDate: new Date(2020, 1, 22) })
  await actions.createPost({ aDate: new Date(2020, 1, 23) })

  const { collection: response } = await actions.listPostsNotEqualDate({ 
      whereArg: new Date(2020, 1, 21)
  })

  expect(response.length).toEqual(2)
})

test('List Where filters - using after operator on timestamp - filters correctly', async () => {
  await actions.createPost({ aTimestamp: new Date(2020, 1, 21, 1, 0, 0) })
  await actions.createPost({ aTimestamp: new Date(2020, 1, 22, 2, 30, 0) })
  await actions.createPost({ aTimestamp: new Date(2020, 1, 23, 4, 0, 0) })

  const { collection: response } = await actions.listPostsAfterTimestamp({ 
      whereArg: new Date(2020, 1, 21, 1, 30, 0)
  })

  expect(response.length).toEqual(2)
})

test('List Where filters - using before operator on timestamp - filters correctly', async () => {
  await actions.createPost({ aTimestamp: new Date(2020, 1, 21, 1, 0, 0) })
  await actions.createPost({ aTimestamp: new Date(2020, 1, 22, 2, 30, 0) })
  await actions.createPost({ aTimestamp: new Date(2020, 1, 23, 4, 0, 0) })

  const { collection: response } = await actions.beforePostsBeforeTimestamp({ 
      whereArg: new Date(2020, 1, 21, 1, 30, 0)
  })

  expect(response.length).toEqual(1)
})

