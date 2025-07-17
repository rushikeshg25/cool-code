"use client";
import React from "react";
import { motion } from "motion/react";

/**
 * Note: This component is optimized for and looks best in dark mode.
 */

export function BackgroundWithFullVideo() {
  return (
    <div className="flex h-screen flex-col items-center justify-center">
      <Background />
      <Content />
    </div>
  );
}

const Background = () => {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 1, ease: "easeOut" }}
      className="pointer-events-none absolute inset-0 z-0 h-full w-full overflow-hidden dark:opacity-20"
    >
      <video
        autoPlay
        muted
        loop
        playsInline
        className="h-full w-full object-cover dark:[mask-image:radial-gradient(circle_at_center,white,transparent)]"
      >
        <source
          src="https://assets.aceternity.com/background-demo.mp4"
          type="video/mp4"
        />
      </video>
    </motion.div>
  );
};

const Content = () => {
  return (
    <div className="relative z-10">
      <div className="pointer-events-none absolute inset-0 h-full w-full rounded-full bg-black blur-2xl"></div>
      <h1 className="z-2 relative text-center font-sans text-2xl font-bold text-white md:text-5xl lg:text-7xl">
        The best <ColourfulText text="components" /> <br /> you will ever find
      </h1>
    </div>
  );
};

export function ColourfulText({ text }: { text: string }) {
  const colors = [
    "rgb(131, 179, 32)",
    "rgb(47, 195, 106)",
    "rgb(42, 169, 210)",
    "rgb(4, 112, 202)",
    "rgb(107, 10, 255)",
    "rgb(183, 0, 218)",
    "rgb(218, 0, 171)",
    "rgb(230, 64, 92)",
    "rgb(232, 98, 63)",
    "rgb(249, 129, 47)",
  ];

  const [currentColors, setCurrentColors] = React.useState(colors);
  const [count, setCount] = React.useState(0);

  React.useEffect(() => {
    const interval = setInterval(() => {
      const shuffled = [...colors].sort(() => Math.random() - 0.5);
      setCurrentColors(shuffled);
      setCount((prev) => prev + 1);
    }, 5000);

    return () => clearInterval(interval);
  }, []);

  return text.split("").map((char, index) => (
    <motion.span
      key={`${char}-${count}-${index}`}
      initial={{
        y: 0,
      }}
      animate={{
        color: currentColors[index % currentColors.length],
        y: [0, -3, 0],
        scale: [1, 1.01, 1],
        filter: ["blur(0px)", `blur(5px)`, "blur(0px)"],
        opacity: [1, 0.8, 1],
      }}
      transition={{
        duration: 0.5,
        delay: index * 0.05,
      }}
      className="inline-block whitespace-pre font-sans tracking-tight"
    >
      {char}
    </motion.span>
  ));
}
