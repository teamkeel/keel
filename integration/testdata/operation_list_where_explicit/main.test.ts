import { test, expect, actions, Post, logger } from '@teamkeel/testing'
import { LogLevel } from '@teamkeel/sdk'

test('List Where filters - using equal operator (string) - filters correctly', async () => {
  await Post.create({ title: 'Fred' })
  await Post.create({ title: 'NotFred' })

  const { collection: response } = await actions.listPostsEqualString({ 
      whereArg: "Fred" 
    })

  expect(response.length).toEqual(1)
  expect(response[0].title).toEqual('Fred')
})


test('List Where filters - using not equal operator (string) - filters correctly', async () => {
  await Post.create({ title: 'Fred' })
  await Post.create({ title: 'NotFred' })

  const { collection: response } = await actions.listPostsNotEqualString({ 
      where: {
        whereArg: "Fred"
      },
      pageInfo: {
        
      }
    })

  expect(response.length).toEqual(1)
  expect(response[0].title).toEqual('NotFred')
})

// todo get Date and Time versions of these tests working as above.
// Currently the generated Post.create() method, requires that we pass in native JS Date objects -
// but these do not make it through to the actions Go code - that expects to receive graphql style
// date/time structures with fields for seconds/year etc.



// test('List Where filters - using equal operator (date) - filters correctly', async () => {
//   await Post.create({ aDate: new Date(2020, 1, 21).toISOString() })
//   await Post.create({ aDate: new Date(2020, 1, 22) .toISOString() })

//   const { collection: response } = await actions.listPostsEqualDate({ 
//       whereArg: { aDate: new Date(2020, 1, 21)}
//   })

//   console.log(response)
//   expect(response.length).toEqual(1)
//   // expect(response[0].aDate.day).toEqual(22)
// })
