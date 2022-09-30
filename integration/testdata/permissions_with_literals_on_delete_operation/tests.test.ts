import { test, expect, actions, Post } from '@teamkeel/testing'

test('text permission successful', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPostWithTitle({ title: "hello" })

  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .deletePostTextPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, false)
})

test('text permission failed', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPostWithTitle({ title: "not hello" })

  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .deletePostTextPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('number permission successful', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPostWithViews({ views: 5 })

  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .deletePostNumberPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, false)
})

test('number permission failed', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPostWithViews({ views: 500 })

  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .deletePostNumberPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('boolean permission successful', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPostWithActive({ active: true })

  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .deletePostBooleanPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, false)
})

test('boolean permission failed', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPostWithActive({ active: false })

  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .deletePostBooleanPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('text not null permission successful', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPostWithTitle({ title: "hello" })

  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .deletePostTextNotNullPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, false)
})

test('text not null permission failed', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPostWithTitle({ title: null })

  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .deletePostTextNotNullPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('text not null from default permission failed', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPost({})

  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .deletePostTextNotNullPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('number not null permission successful', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPostWithViews({ views: 5 })

  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .deletePostNumberNotNullPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, false)
})

test('number not null permission failed', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPostWithViews({ views: null })

  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .deletePostNumberNotNullPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('number not null from default permission failed', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPost({})

  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .deletePostNumberNotNullPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('boolean not null permission successful', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPostWithActive({ active: true })

  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .deletePostBooleanNotNullPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, false)
})

test('boolean not null permission failed', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPostWithActive({ active: null })

  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .deletePostBooleanNotNullPermission({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('boolean not null from default permission failed', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPost({})

  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
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