import { test, expect, Actions, Person } from '@teamkeel/testing'

test('creating a person', async () => {
  const result = await Actions.createPerson({ title: 'foo' })
  expect.equal(result.title, 'foo')
})

test('fetching a person by id', async () => {
  const person = await Person.create({ title: 'bar' })
  const res = await Actions.getPerson({ id: person.id })

  expect.equal(res.id, person.id)
  expect.equal(res.title, person.title)
})
