import { PrismaClient } from '@prisma/client';
import { NextResponse } from 'next/server';

const prisma = new PrismaClient();

export async function POST(request: Request) {
  try {
    const { title, artist, albumArt } = await request.json();

    if (!title || !artist || !albumArt) {
      return new NextResponse(JSON.stringify({ error: 'Title, artist, and albumArt are required' }), { status: 400 });
    }

    const newAlbum = await prisma.madeForYouAlbum.create({
      data: {
        title,
        artist,
        albumArt,
      },
    });

    return new NextResponse(JSON.stringify(newAlbum), { status: 201 });
  } catch (error) {
    console.error('Error recording Made For You album:', error);
    return new NextResponse(JSON.stringify({ error: 'Failed to record Made For You album' }), { status: 500 });
  }
}

export async function GET() {
  try {
    const albums = await prisma.madeForYouAlbum.findMany({
      orderBy: {
        createdAt: 'desc',
      },
    });
    return new NextResponse(JSON.stringify(albums), { status: 200 });
  } catch (error) {
    console.error('Error fetching Made For You albums:', error);
    return new NextResponse(JSON.stringify({ error: 'Failed to fetch Made For You albums' }), { status: 500 });
  }
}