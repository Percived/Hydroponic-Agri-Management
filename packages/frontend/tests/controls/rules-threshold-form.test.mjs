import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

const currentFileDir = dirname(fileURLToPath(import.meta.url))
const projectRoot = resolve(currentFileDir, '..', '..')
const rulesViewPath = resolve(projectRoot, 'src', 'views', 'controls', 'rules.vue')

test('新增策略表单的阈值条件区不应被额外 template 包裹', () => {
  const content = readFileSync(rulesViewPath, 'utf8')

  assert.match(content, /<template v-if="formData\.policy_type === 'THRESHOLD'">/)
  assert.doesNotMatch(
    content,
    /<template v-if="formData\.policy_type === 'THRESHOLD'">[\s\S]*?<el-divider>策略条件<\/el-divider>\s*<template>/
  )
})
