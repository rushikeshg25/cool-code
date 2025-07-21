-- CreateTable
CREATE TABLE "RecentlyPlayedSong" (
    "id" TEXT NOT NULL PRIMARY KEY,
    "title" TEXT NOT NULL,
    "artist" TEXT NOT NULL,
    "album" TEXT,
    "playedAt" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
