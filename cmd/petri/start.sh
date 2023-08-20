#!/bin/sh
npx prisma migrate deploy
npx remix build
NODE_OPTIONS=--max_old_space_size=4096 NODE_ENV=production npx ts-node --require tsconfig-paths/register server/index.tst
