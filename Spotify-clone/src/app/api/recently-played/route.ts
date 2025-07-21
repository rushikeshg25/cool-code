import { PrismaClient } from '@prisma/client';
import { NextResponse } from 'next/server';

const prisma = new PrismaClient();

export async function POST(request: Request) {
  try {
    const { title, artist, album } = await request.json();

    if (!title || !artist) {
      return new NextResponse(JSON.stringify({ error: 'Title and artist are required' }), { status: 400 });
    }

    const newSong = await prisma.recentlyPlayedSong.create({
      data: {
        title,
        artist,
        album,
      },
    });

    return new NextResponse(JSON.stringify(newSong), { status: 201 });
  } catch (error) {
    console.error('Error recording song:', error);
    return new NextResponse(JSON.stringify({ error: 'Failed to record song' }), { status: 500 });
  }
}

export async function GET() {
  try {
    const songs = await prisma.recentlyPlayedSong.findMany({
      orderBy: {
        playedAt: 'desc',
      },
    });
    return new NextResponse(JSON.stringify(songs), { status: 200 });
  } catch (error) {
    console.error('Error fetching songs:', error);
    return new NextResponse(JSON.stringify({ error: 'Failed to fetch songs' }), { status: 500 });
  }
}