import { test, expect, Actions, Person, logger } from '@teamkeel/testing'

test('creating a person', async () => {
  const { result } = await Actions.createPerson({ name: 'foo', gender: 'female', nINumber: '282' })
  expect.equal(result.name, 'foo')
})

test('fetching a person by id', async () => {
  const person = await Person.create({ name: 'bar', gender: 'male', nINumber: '123' })
  const { result: res } = await Actions.getPerson({ id: person.id })

  expect.equal(res.id, person.id)
  expect.equal(res.name, person.name)
})

test('fetching person by unique NINumber field', async () => {
  const person = await Person.create({ name: 'bar', gender: 'male', nINumber: '333' })

  const { result: res } = await Actions.getPerson({ nINumber: '333' })

  expect.equal(res.id, person.id)
})
