-- CreateTable
CREATE TABLE "User" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "email" TEXT NOT NULL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL
);

-- CreateTable
CREATE TABLE "Password" (
    "hash" TEXT NOT NULL,
    "userId" TEXT NOT NULL,
    CONSTRAINT "Password_userId_fkey" FOREIGN KEY ("userId") REFERENCES "User" ("id") ON DELETE CASCADE ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "Net" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "name" TEXT NOT NULL,
    "description" TEXT NOT NULL,
    "createdAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updatedAt" DATETIME NOT NULL,
    "authorID" TEXT NOT NULL,
    CONSTRAINT "Net_authorID_fkey" FOREIGN KEY ("authorID") REFERENCES "User" ("id") ON DELETE RESTRICT ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "Place" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "netID" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    "bound" INTEGER NOT NULL,
    CONSTRAINT "Place_netID_fkey" FOREIGN KEY ("netID") REFERENCES "Net" ("id") ON DELETE RESTRICT ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "Transition" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "netID" TEXT NOT NULL,
    "name" TEXT NOT NULL,
    CONSTRAINT "Transition_netID_fkey" FOREIGN KEY ("netID") REFERENCES "Net" ("id") ON DELETE RESTRICT ON UPDATE CASCADE
);

-- CreateTable
CREATE TABLE "Arc" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "netID" TEXT NOT NULL,
    "fromPlace" BOOLEAN NOT NULL,
    "placeID" TEXT NOT NULL,
    "transitionID" TEXT NOT NULL,
    CONSTRAINT "Arc_netID_fkey" FOREIGN KEY ("netID") REFERENCES "Net" ("id") ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT "Arc_placeID_fkey" FOREIGN KEY ("placeID") REFERENCES "Place" ("id") ON DELETE RESTRICT ON UPDATE CASCADE,
    CONSTRAINT "Arc_transitionID_fkey" FOREIGN KEY ("transitionID") REFERENCES "Transition" ("id") ON DELETE RESTRICT ON UPDATE CASCADE
);

-- CreateIndex
CREATE UNIQUE INDEX "User_email_key" ON "User"("email");

-- CreateIndex
CREATE UNIQUE INDEX "Password_userId_key" ON "Password"("userId");
