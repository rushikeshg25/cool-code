"use client";
import { cn } from "@/lib/utils";
import React, { useRef } from "react";
import { motion, useInView, useSpring, useTransform } from "motion/react";
import { useEffect } from "react";

export function StatsWithNumberTicker() {
  const items = [
    {
      description:
        "Years in the business trying to build this business from the ground up.",
      value: 10,
      subtext: "trillion",
    },
    {
      description: "People Fought trying to establish our brand.",
      value: 100,
      subtext: "million",
    },
    {
      description:
        "Fight Club Attendance so that everyone knows the first rule.",
      value: 25724,
      subtext: "million",
    },
    {
      description: "People trying to get access but guess what? they can't.",
      value: 69420,
      subtext: "million",
    },
  ];
  return (
    <section className="group/container relative mx-auto w-full max-w-7xl overflow-hidden rounded-3xl p-10 py-20">
      <div className="relative z-20">
        <h2 className="text-center text-xl font-bold text-neutral-700 md:text-3xl dark:text-neutral-100">
          Trusted by fighters all over the world
        </h2>
        <p className="mx-auto mt-4 max-w-2xl text-center text-sm text-neutral-800 md:text-base dark:text-neutral-200">
          We are a team of experienced fighters and boxers who are passionate
          about helping you grow your business.
        </p>
        <div className="mx-auto mt-10 grid max-w-7xl grid-cols-1 gap-10 sm:grid-cols-2 md:grid-cols-4">
          {items.map((item, index) => (
            <motion.div
              initial={{ y: 20, opacity: 0, filter: "blur(4px)" }}
              animate={{ y: 0, opacity: 1, filter: "blur(0px)" }}
              transition={{ duration: 0.5, delay: index * 0.1 }}
              key={"card" + index}
              className={cn("group/card relative overflow-hidden rounded-lg")}
            >
              <div className="flex items-center gap-2">
                <p className="text-3xl font-bold text-neutral-700 dark:text-neutral-200">
                  <AnimatedNumber value={item.value} />
                </p>
              </div>
              <p className="mt-4 text-balance text-base text-neutral-600 dark:text-neutral-300">
                {item.description}
              </p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

function AnimatedNumber({
  value,
  initial = 0,
}: {
  value: number;
  initial?: number;
}) {
  const ref = useRef(null);
  const isInView = useInView(ref);

  const spring = useSpring(initial, { mass: 0.8, stiffness: 75, damping: 15 });
  const display = useTransform(spring, (current) =>
    Math.round(current).toLocaleString()
  );

  useEffect(() => {
    if (isInView) {
      spring.set(value);
    } else {
      spring.set(initial);
    }
  }, [isInView, spring, value, initial]);

  return <motion.span ref={ref}>{display}</motion.span>;
}
