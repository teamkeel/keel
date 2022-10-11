import { test, expect, actions } from '@teamkeel/testing'

test('boolean literal - matching value - is authorized', async () => {
  expect(
    await actions
    .createWithTrueValuePermission({ title: "hello" })
  ).notToHaveAuthorizationError()
})

test('input arg - matching value - is authorized', async () => {
  expect(
    await actions
    .createWithInputArgPermission({ title: "123" })
  ).notToHaveAuthorizationError()
})

test('input arg - non matching value - is not authorized', async () => {
  expect(
    await actions
    .createWithInputArgPermission({ title: "321" })
  ).toHaveAuthorizationError()
})