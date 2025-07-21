import { PrismaClient } from '@prisma/client';

const prisma = new PrismaClient();

async function main() {
  console.log('Start seeding...');

  // Seed MadeForYouAlbum
  await prisma.madeForYouAlbum.createMany({
    data: [
      {
        title: 'Chill Vibes',
        artist: 'Various Artists',
        albumArt: 'https://misc.scdn.co/made-for-you/alt-text/chill_alt_text.jpg',
      },
      {
        title: 'Workout Mix',
        artist: 'Fitness Beats',
        albumArt: 'https://misc.scdn.co/made-for-you/alt-text/workout_alt_text.jpg',
      },
      {
        title: 'Focus Flow',
        artist: 'Ambient Sounds',
        albumArt: 'https://misc.scdn.co/made-for-you/alt-text/focus_alt_text.jpg',
      },
    ],
  });

  // Seed PopularAlbum
  await prisma.popularAlbum.createMany({
    data: [
      {
        title: 'The Best Of',
        artist: 'Classic Rock',
        albumArt: 'https://i.scdn.co/image/ab67616d0000b273e2e2e2e2e2e2e2e2e2e2e2e2',
      },
      {
        title: 'Pop Hits 2024',
        artist: 'Various Pop Artists',
        albumArt: 'https://i.scdn.co/image/ab67616d0000b273f3f3f3f3f3f3f3f3f3f3f3f3',
      },
      {
        title: 'Indie Discoveries',
        artist: 'New Indie Bands',
        albumArt: 'https://i.scdn.co/image/ab67616d0000b273g4g4g4g4g4g4g4g4g4g4g4g4',
      },
    ],
  });

  console.log('Seeding finished.');
}

main()
  .catch((e) => {
    console.error(e);
    process.exit(1);
  })
  .finally(async () => {
    await prisma.$disconnect();
  });
