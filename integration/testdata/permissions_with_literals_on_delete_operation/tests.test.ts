import { test, expect, actions, Post } from '@teamkeel/testing'

test('text permission successful', async () => {
  const { object: post } = await actions
    .createPostWithTitle({ title: "hello" })

  const { errors } = await actions
    .deletePostTextPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, false)
})

test('text permission failed', async () => {
  const { object: post } = await actions
    .createPostWithTitle({ title: "not hello" })

  const { errors } = await actions
    .deletePostTextPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('number permission successful', async () => {
  const { object: post } = await actions
    .createPostWithViews({ views: 5 })

  const { errors } = await actions
    .deletePostNumberPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, false)
})

test('number permission failed', async () => {
  const { object: post } = await actions
    .createPostWithViews({ views: 500 })

  const { errors } = await actions
    .deletePostNumberPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('boolean permission successful', async () => {
  const { object: post } = await actions
    .createPostWithActive({ active: true })

  const { errors } = await actions
    .deletePostBooleanPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, false)
})

test('boolean permission failed', async () => {
  const { object: post } = await actions
    .createPostWithActive({ active: false })

  const { errors } = await actions
    .deletePostBooleanPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('text not null permission successful', async () => {
  const { object: post } = await actions
    .createPostWithTitle({ title: "hello" })

  const { errors } = await actions
    .deletePostTextNotNullPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, false)
})

test('text not null permission failed', async () => {
  const { object: post } = await actions
    .createPostWithTitle({ title: null })

  const { errors } = await actions
    .deletePostTextNotNullPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('text not null from default permission failed', async () => {
  const { object: post } = await actions
    .createPost({})

  const { errors } = await actions
    .deletePostTextNotNullPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('number not null permission successful', async () => {
  const { object: post } = await actions
    .createPostWithViews({ views: 5 })

  const { errors } = await actions
    .deletePostNumberNotNullPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, false)
})

test('number not null permission failed', async () => {
  const { object: post } = await actions
    .createPostWithViews({ views: null })

  const { errors } = await actions
    .deletePostNumberNotNullPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('number not null from default permission failed', async () => {
  const { object: post } = await actions
    .createPost({})

  const { errors } = await actions
    .deletePostNumberNotNullPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('boolean not null permission successful', async () => {
  const { object: post } = await actions
    .createPostWithActive({ active: true })

  const { errors } = await actions
    .deletePostBooleanNotNullPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, false)
})

test('boolean not null permission failed', async () => {
  const { object: post } = await actions
    .createPostWithActive({ active: null })

  const { errors } = await actions
    .deletePostBooleanNotNullPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('boolean not null from default permission failed', async () => {
  const { object: post } = await actions
    .createPost({})

  const { errors } = await actions
    .deletePostBooleanNotNullPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

function hasAuthorizationError(errors?): boolean {
  if (errors == null)
    return false;

  var hasError = false
   errors.forEach(function(error) {
    if(error.message == 'not authorized to access this operation') {
      hasError = true
    }
  });
  
  return hasError;
}