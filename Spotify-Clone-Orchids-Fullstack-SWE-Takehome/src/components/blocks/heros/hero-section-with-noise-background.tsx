"use client";
import React, { useEffect, useState } from "react";
import { motion } from "motion/react";
export function HeroSectionWithNoiseBackground() {
  return (
    <div className="relative flex w-full items-center justify-center overflow-hidden bg-white px-4 py-20 md:py-40 dark:bg-black">
      <Background />

      <div className="relative z-10 mx-auto w-full max-w-7xl">
        <Badge text="We raised $69M pre seed" />
        <motion.h2
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3 }}
          className="text-center text-3xl font-bold tracking-tight text-neutral-800 md:text-4xl lg:text-7xl dark:text-neutral-100"
        >
          Write fast with <br /> accurate precision.
        </motion.h2>
        <motion.p
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3, delay: 0.2 }}
          className="mx-auto mt-4 max-w-lg text-center text-sm text-neutral-600 md:text-xl dark:text-neutral-400"
        >
          Our state of the art tool is a tool that allows you to write copy
          instantly.
        </motion.p>

        <motion.div
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3, delay: 0.2 }}
          className="flex flex-col items-center gap-4 md:flex-row"
        >
          <div className="mt-8 flex w-full flex-col justify-center gap-4 sm:flex-row">
            <button className="rounded-lg bg-neutral-900 px-6 py-3 text-base font-medium text-white hover:bg-neutral-800 dark:bg-neutral-100 dark:text-neutral-900 dark:hover:bg-neutral-200">
              Download for Linux
            </button>
            <button className="dark:hover:bg-fuschia-600 darhk:text-white rounded-lg bg-rose-500 px-6 py-3 text-base font-medium text-white shadow-[0_1px_0_0_#0ea5e9,0_-1px_0_0_#0ea5eh9] hover:bg-rose-600 dark:bg-rose-500 dark:shadow-[0_1px_0_0_#0369a1,0_-1px_0_0_#0369ah1]">
              Download for Windows
            </button>
          </div>
        </motion.div>

        <div className="z-40 mt-12 flex w-full justify-center bg-white dark:bg-black">
          <motion.div
            initial={{ opacity: 0, y: 20, filter: "blur(10px)" }}
            animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
            transition={{ duration: 0.3, delay: 0.3 }}
            className="relative w-full overflow-hidden rounded-xl shadow-2xl [mask-image:linear-gradient(to_bottom,white,white_40%,transparent)]"
          >
            <img
              src="https://assets.aceternity.com/linear-demo.webp"
              alt="Product screenshot"
              className="h-auto w-full object-cover"
              width={1024}
              height={576}
            />
            <div className="absolute inset-0 bg-gradient-to-t from-black/20 to-transparent opacity-50"></div>
          </motion.div>
        </div>
      </div>
    </div>
  );
}

const Badge = ({ text }: { text: string }) => {
  return (
    <motion.div
      initial={{ opacity: 0, y: 10 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3 }}
      className="relative mx-auto mb-6 flex w-fit items-center justify-center overflow-hidden rounded-full p-px"
    >
      <motion.div
        className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 rounded-full bg-gradient-to-r from-transparent via-transparent to-blue-500"
        animate={{ rotate: 360 }}
        transition={{ duration: 3, repeat: Infinity, ease: "linear" }}
        style={{ width: "300px", height: "20px" }}
      />
      <div className="relative z-10 rounded-full bg-neutral-100 px-4 py-2 text-sm font-medium text-neutral-800 dark:bg-neutral-800 dark:text-neutral-100">
        {text}
      </div>
    </motion.div>
  );
};

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
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 1 }}
      className="absolute inset-0 z-0 flex [mask-image:radial-gradient(circle_at_center,white_0%,white_30%,transparent_70%)]"
    >
      <Noise />
      {strips.map((index) => (
        <div
          key={index}
          className="h-full w-20 bg-gradient-to-r from-neutral-100 to-white shadow-[2px_0px_0px_0px_var(--color-neutral-400)] dark:from-neutral-900 dark:to-neutral-950 dark:shadow-[2px_0px_0px_0px_var(--color-neutral-800)]"
        />
      ))}
    </motion.div>
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
