"use client";
import { cn } from "@/lib/utils";
import { AnimationProps, motion } from "motion/react";
import React from "react";

export function TextAnimationBlurFadeInDemo() {
  return (
    <div className="mx-auto max-w-4xl p-4">
      <Text className="mb-4 text-center text-2xl font-bold tracking-tight md:text-7xl">
        Animations that stand out from the rest.
      </Text>
      <Text
        className="mx-auto max-w-2xl text-center text-xl font-normal text-neutral-600 dark:text-neutral-600"
        delay={0.2}
      >
        Appearing text sets aside the text from the rest, it is a great way to
        make the text more engaging and interesting.
      </Text>

      <div className="mt-8 flex flex-col justify-center gap-4 md:flex-row">
        <motion.button
          className="rounded-full bg-black px-6 py-3 text-white transition-colors hover:bg-black/80 dark:bg-white dark:text-black dark:hover:bg-white/80"
          initial={{ opacity: 0, y: 10 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3, delay: 0.8 }}
        >
          Get Started
        </motion.button>
        <motion.button
          className="rounded-full border border-black px-6 py-3 text-black transition-colors hover:bg-black/5 dark:border-white dark:text-white dark:hover:bg-white/5"
          initial={{ opacity: 0, y: 10 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3, delay: 1 }}
        >
          Learn More
        </motion.button>
      </div>
    </div>
  );
}

const Text = ({
  children,
  className,
  delay = 0,
  ...animationProps
}: {
  children: string;
  className?: string;
  delay?: number;
} & AnimationProps) => {
  return (
    <motion.p
      {...animationProps}
      className={cn("text-4xl font-medium", className)}
    >
      {children.split(" ").map((word, index) => (
        <motion.span
          key={`word-${index}-${word}`}
          initial={{
            opacity: 0,
            filter: "blur(10px)",
            y: 10,
          }}
          whileInView={{
            opacity: 1,
            filter: "blur(0px)",
            y: 0,
          }}
          transition={{
            duration: 0.2,
            delay: delay + index * 0.02,
          }}
          className="inline-block"
        >
          {word}&nbsp;
        </motion.span>
      ))}
    </motion.p>
  );
};
