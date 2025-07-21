"use client";
import { cn } from "@/lib/utils";
import React, { useEffect, useState, useRef } from "react";
import { motion } from "motion/react";

export function BackgroundLines() {
  return (
    <div className="flex h-screen flex-col items-center justify-center p-4">
      <Content />
      <ShootingStars
        minSpeed={10}
        maxSpeed={30}
        minDelay={300}
        maxDelay={1000}
        starColor="#9E00FF"
        trailColor="#2EB9DF"
        starWidth={10}
        starHeight={1}
      />
      <Circles />
    </div>
  );
}

const Circles = () => {
  return (
    <div className="pointer-events-none absolute inset-0 flex h-full w-full items-center justify-center overflow-hidden">
      <motion.div
        className="pointer-events-none absolute inset-0 m-auto h-80 w-80 rounded-full bg-gradient-to-b from-neutral-300 to-transparent to-[40%] dark:from-neutral-700"
        animate={{ scale: [1, 1.05, 1] }}
        transition={{
          duration: 1,
          ease: "linear",
          repeat: Infinity,
          repeatDelay: 2,
        }}
      />
      <motion.div
        className="pointer-events-none absolute inset-0 m-auto h-[24rem] w-[24rem] rounded-full bg-gradient-to-b from-neutral-200 to-transparent to-[40%] dark:from-neutral-800"
        animate={{ scale: [1, 1.05, 1] }}
        transition={{
          duration: 1,
          ease: "linear",
          repeat: Infinity,
          delay: 0.3,
          repeatDelay: 2,
        }}
      />
      <motion.div
        className="pointer-events-none absolute inset-0 m-auto h-[28rem] w-[28rem] rounded-full bg-gradient-to-b from-neutral-100 to-transparent to-[40%] dark:from-neutral-900"
        animate={{ scale: [1, 1.05, 1] }}
        transition={{
          duration: 1,
          ease: "linear",
          repeat: Infinity,
          delay: 0.6,
          repeatDelay: 2,
        }}
      />
    </div>
  );
};

const Content = () => {
  return (
    <div className="relative z-10">
      <h1 className="text-balance mx-auto max-w-2xl text-center text-3xl font-bold text-black dark:text-white md:text-5xl">
        Web apps that make you feel like you&apos;re in the future
      </h1>
      <p className="text-balance mx-auto mt-4 max-w-2xl text-center text-base text-neutral-800 dark:text-neutral-200">
        We are a team of developers who are passionate about creating web apps
        that make you feel like you&apos;re in the future.
      </p>
      <div className="mt-8 flex flex-col items-center justify-center gap-4 sm:flex-row">
        <button className="w-40 rounded-lg bg-gradient-to-b from-neutral-950 to-neutral-800 px-4 py-2 text-sm text-white">
          Sign up
        </button>
        <button className="w-40 rounded-lg bg-gradient-to-b from-neutral-100 to-neutral-50 px-4 py-2 text-sm text-black">
          Register now
        </button>
      </div>
    </div>
  );
};

interface ShootingStar {
  id: number;
  x: number;
  y: number;
  angle: number;
  scale: number;
  speed: number;
  distance: number;
}

interface ShootingStarsProps {
  minSpeed?: number;
  maxSpeed?: number;
  minDelay?: number;
  maxDelay?: number;
  starColor?: string;
  trailColor?: string;
  starWidth?: number;
  starHeight?: number;
  className?: string;
}

const getRandomStartPoint = () => {
  const side = Math.floor(Math.random() * 4);
  const offset = Math.random() * window.innerWidth;

  switch (side) {
    case 0:
      return { x: offset, y: 0, angle: 45 };
    case 1:
      return { x: window.innerWidth, y: offset, angle: 135 };
    case 2:
      return { x: offset, y: window.innerHeight, angle: 225 };
    case 3:
      return { x: 0, y: offset, angle: 315 };
    default:
      return { x: 0, y: 0, angle: 45 };
  }
};
export const ShootingStars: React.FC<ShootingStarsProps> = ({
  minSpeed = 10,
  maxSpeed = 30,
  minDelay = 1200,
  maxDelay = 4200,
  starColor = "#9E00FF",
  trailColor = "#2EB9DF",
  starWidth = 10,
  starHeight = 1,
  className,
}) => {
  const [star, setStar] = useState<ShootingStar | null>(null);
  const svgRef = useRef<SVGSVGElement>(null);

  useEffect(() => {
    const createStar = () => {
      const { x, y, angle } = getRandomStartPoint();
      const newStar: ShootingStar = {
        id: Date.now(),
        x,
        y,
        angle,
        scale: 1,
        speed: Math.random() * (maxSpeed - minSpeed) + minSpeed,
        distance: 0,
      };
      setStar(newStar);

      const randomDelay = Math.random() * (maxDelay - minDelay) + minDelay;
      setTimeout(createStar, randomDelay);
    };

    createStar();

    return () => {};
  }, [minSpeed, maxSpeed, minDelay, maxDelay]);

  useEffect(() => {
    const moveStar = () => {
      if (star) {
        setStar((prevStar) => {
          if (!prevStar) return null;
          const newX =
            prevStar.x +
            prevStar.speed * Math.cos((prevStar.angle * Math.PI) / 180);
          const newY =
            prevStar.y +
            prevStar.speed * Math.sin((prevStar.angle * Math.PI) / 180);
          const newDistance = prevStar.distance + prevStar.speed;
          const newScale = 1 + newDistance / 100;
          if (
            newX < -20 ||
            newX > window.innerWidth + 20 ||
            newY < -20 ||
            newY > window.innerHeight + 20
          ) {
            return null;
          }
          return {
            ...prevStar,
            x: newX,
            y: newY,
            distance: newDistance,
            scale: newScale,
          };
        });
      }
    };

    const animationFrame = requestAnimationFrame(moveStar);
    return () => cancelAnimationFrame(animationFrame);
  }, [star]);

  return (
    <svg
      ref={svgRef}
      className={cn("absolute inset-0 h-full w-full", className)}
    >
      {star && (
        <rect
          key={star.id}
          x={star.x}
          y={star.y}
          width={starWidth * star.scale}
          height={starHeight}
          fill="url(#gradient)"
          transform={`rotate(${star.angle}, ${
            star.x + (starWidth * star.scale) / 2
          }, ${star.y + starHeight / 2})`}
        />
      )}
      <defs>
        <linearGradient id="gradient" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" style={{ stopColor: trailColor, stopOpacity: 0 }} />
          <stop
            offset="100%"
            style={{ stopColor: starColor, stopOpacity: 1 }}
          />
        </linearGradient>
      </defs>
    </svg>
  );
};
