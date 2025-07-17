"use client";
import React, { useEffect, useState } from "react";

export function BackgroundNoiseGrid() {
  return (
    <div className="relative flex h-screen w-full items-center justify-center overflow-hidden bg-white dark:bg-black">
      <Background />

      <div className="relative z-10 mx-auto max-w-4xl">
        <h2 className="text-center text-2xl font-bold tracking-tight text-neutral-800 md:text-4xl lg:text-7xl dark:text-neutral-100">
          Instant copywriting with our <br /> state of the art tool.
        </h2>
      </div>
    </div>
  );
}

const Background = () => {
  const [strips, setStrips] = useState<number[]>([]);
  useEffect(() => {
    const calculateStrips = () => {
      const viewportWidth = window.innerWidth;
      const stripWidth = 80;
      const numberOfStrips = Math.ceil(viewportWidth / stripWidth);
      setStrips(Array.from({ length: numberOfStrips }, (_, i) => i));
    };
    calculateStrips();
    window.addEventListener("resize", calculateStrips);
    return () => window.removeEventListener("resize", calculateStrips);
  }, []);

  return (
    <div className="absolute inset-0 z-0 flex">
      <Noise />
      {strips.map((index) => (
        <div
          key={index}
          className="h-full w-20 bg-gradient-to-r from-neutral-100 to-white shadow-[2px_0px_0px_0px_var(--color-neutral-400)] dark:from-neutral-900 dark:to-neutral-950 dark:shadow-[2px_0px_0px_0px_var(--color-neutral-800)]"
        />
      ))}
    </div>
  );
};

const Noise = () => {
  return (
    <div
      className="absolute inset-0 h-full w-full scale-[1.2] transform opacity-[0.05] [mask-image:radial-gradient(#fff,transparent,75%)]"
      style={{
        backgroundImage: "url(/noise.webp)",
        backgroundSize: "20%",
      }}
    ></div>
  );
};
