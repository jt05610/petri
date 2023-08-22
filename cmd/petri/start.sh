#!/bin/sh
npx prisma migrate deploy
npx remix build
npx remix-serve ./build