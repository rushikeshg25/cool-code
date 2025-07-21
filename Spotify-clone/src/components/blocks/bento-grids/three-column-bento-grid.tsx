"use client";

import { cn } from "@/lib/utils";
import React, { useEffect, useState, useRef, useId } from "react";
import { motion } from "motion/react";
import Image from "next/image";
import { IconUpload } from "@tabler/icons-react";

export function ThreeColumnBentoGrid() {
  return (
    <div className="mx-auto my-20 w-full max-w-7xl px-4 md:px-8">
      <h2 className="text-bold text-neutral-8000 font-sans text-xl font-bold tracking-tight md:text-4xl dark:text-neutral-100">
        The perfect ATS score platform
      </h2>
      <p className="mt-4 max-w-lg text-sm text-neutral-600 dark:text-neutral-400">
        Get the most accurate details on your candidate with our state of the
        art, blazing fast, absolutely zero fact based ATS score generator.
      </p>
      <div className="mt-20 grid grid-flow-col grid-cols-1 grid-rows-6 gap-2 md:grid-cols-2 md:grid-rows-3 xl:grid-cols-3 xl:grid-rows-2">
        <Card className="row-span-2">
          <CardContent>
            <CardTitle>Generate scores based on pictures</CardTitle>
            <CardDescription>
              Rate your candidate&apos;s looks. As real as ATS scores.
            </CardDescription>
          </CardContent>
          <CardSkeletonBody>
            <SkeletonOne />
          </CardSkeletonBody>
        </Card>
        <Card className="overflow-hidden">
          <CardContent>
            <CardTitle>Track progress</CardTitle>
            <CardDescription>
              Track every step of the candidate&apos;s journey, from initial
              application to rejected appraisal.
            </CardDescription>
          </CardContent>
          <CardSkeletonBody>
            <SkeletonTwo />
          </CardSkeletonBody>
        </Card>
        <Card>
          <CardContent>
            <CardTitle>Schedule interviews seamlessly</CardTitle>
            <CardDescription>
              Ask about DSA or Dev, we don&apos;t care.
            </CardDescription>
          </CardContent>
          <CardSkeletonBody className="">
            <SkeletonThree />
          </CardSkeletonBody>
        </Card>
        <Card className="row-span-2">
          <CardContent>
            <CardTitle>Easy upload resumes manually</CardTitle>
            <CardDescription>
              One click OR drag and drop candidate&apos;s resumes.
            </CardDescription>
          </CardContent>
          <CardSkeletonBody className="h-full max-h-full overflow-hidden">
            <SkeletonFour />
          </CardSkeletonBody>
        </Card>
      </div>
    </div>
  );
}

// Skeletons

const SkeletonOne = () => {
  const avatars = [
    {
      src: "https://assets.aceternity.com/pro/headshots/headshot-1.png",
      score: 69,
    },
    {
      src: "https://assets.aceternity.com/pro/headshots/headshot-2.png",
      score: 94,
    },
    {
      src: "https://assets.aceternity.com/pro/headshots/headshot-3.png",
      score: 92,
    },
    {
      src: "https://assets.aceternity.com/pro/headshots/headshot-4.png",
      score: 98,
    },
  ];

  const [active, setActive] = useState(avatars[3]);

  const intervalTime = 2000; // 2 seconds

  useEffect(() => {
    const interval = setInterval(() => {
      setActive((prev) => {
        const currentIndex = avatars.indexOf(prev);
        const nextIndex = (currentIndex + 1) % avatars.length;
        return avatars[nextIndex];
      });
    }, intervalTime);

    return () => clearInterval(interval);
  }, []);

  const Highlighter = () => {
    return (
      <motion.div layoutId="highlighter" className="absolute inset-0">
        <div className="absolute -left-px -top-px h-4 w-4 rounded-tl-lg border-l-2 border-t-2 border-blue-500 bg-transparent"></div>
        <div className="absolute -right-px -top-px h-4 w-4 rounded-tr-lg border-r-2 border-t-2 border-blue-500 bg-transparent"></div>
        <div className="absolute -bottom-px -left-px h-4 w-4 rounded-bl-lg border-b-2 border-l-2 border-blue-500 bg-transparent"></div>
        <div className="absolute -bottom-px -right-px h-4 w-4 rounded-br-lg border-b-2 border-r-2 border-blue-500 bg-transparent"></div>
      </motion.div>
    );
  };

  const Score = () => {
    return (
      <motion.span
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: 20 }}
        transition={{
          type: "spring",
          stiffness: 260,
          damping: 20,
          delay: 1,
          repeat: 0,
        }}
        className="absolute inset-x-0 bottom-4 m-auto h-fit w-fit rounded-md border border-neutral-100 bg-white px-2 py-1 text-xs text-black"
      >
        score <span className="font-bold">{active.score}</span>
      </motion.span>
    );
  };
  return (
    <div className="p-6">
      <div className="grid grid-cols-2 justify-center gap-4">
        {avatars.map((avatar, index) => (
          <motion.div
            key={`avatar-${index}-avatar-skeleton-one`}
            className="relative"
            animate={{
              opacity: active.src === avatar.src ? 1 : 0.5,
              filter: active.src === avatar.src ? "none" : "grayscale(100%)",
              scale: active.src === avatar.src ? 0.95 : 1,
            }}
            transition={{ duration: 1 }}
          >
            {active.src === avatar.src && <Highlighter />}
            {active.src === avatar.src && <Score />}
            <Image
              key={`avatar-${index}`}
              src={avatar.src}
              alt="avatar"
              width={100}
              height={140}
              className="h-[200px] w-full rounded-lg object-cover object-top"
            />
          </motion.div>
        ))}
      </div>
    </div>
  );
};

export const SkeletonTwo = () => {
  const Cursor = ({
    className,
    textClassName,
    text,
  }: {
    className?: string;
    textClassName?: string;
    text?: string;
  }) => {
    return (
      <div
        className={cn(
          "absolute z-30 h-4 w-4 transition-all duration-200",
          className
        )}
      >
        <svg
          width="19"
          height="19"
          viewBox="0 0 19 19"
          fill="none"
          xmlns="http://www.w3.org/2000/svg"
          className={cn("h-4 w-4 transition duration-200", className)}
        >
          <path
            d="M3.08365 1.18326C2.89589 1.11581 2.70538 1.04739 2.54453 1.00558C2.39192 0.965918 2.09732 0.900171 1.78145 1.00956C1.41932 1.13497 1.13472 1.41956 1.00932 1.78169C0.899927 2.09756 0.965674 2.39216 1.00533 2.54477C1.04714 2.70562 1.11557 2.89613 1.18301 3.0839L5.9571 16.3833C6.04091 16.6168 6.12128 16.8408 6.2006 17.0133C6.26761 17.1591 6.42 17.4781 6.75133 17.6584C7.11364 17.8555 7.54987 17.8612 7.91722 17.6737C8.25317 17.5021 8.41388 17.1873 8.48469 17.0433C8.56852 16.8729 8.65474 16.6511 8.74464 16.4198L10.8936 10.8939L16.4196 8.74489C16.6509 8.655 16.8726 8.56879 17.043 8.48498C17.187 8.41416 17.5018 8.25346 17.6734 7.91751C17.8609 7.55016 17.8552 7.11392 17.6581 6.75162C17.4778 6.42029 17.1589 6.2679 17.0131 6.20089C16.8405 6.12157 16.6165 6.0412 16.383 5.9574L3.08365 1.18326Z"
            fill="var(--blue-900)"
            stroke="var(--blue-500)"
            strokeWidth="1.5"
            strokeLinecap="round"
            strokeLinejoin="round"
          />
        </svg>
        <div
          className={cn(
            "absolute left-3 top-3 whitespace-pre rounded-md p-1 text-[10px] text-neutral-500 transition duration-200",
            textClassName
          )}
        >
          {text ?? "Manu"}
        </div>
      </div>
    );
  };

  const Container = ({
    className,
    children,
  }: {
    className?: string;
    children: React.ReactNode;
  }) => {
    return (
      <div
        className={cn(
          "relative z-20 w-fit rounded-lg border border-neutral-100 p-0.5 shadow-sm dark:border-neutral-600",
          className
        )}
      >
        <div
          className={cn(
            "flex h-10 items-center justify-center rounded-[5px] bg-neutral-100 px-2 text-xs text-neutral-600 shadow-lg dark:bg-[rgba(248,248,248,0.01)] dark:text-neutral-400"
          )}
        >
          {children}
        </div>
      </div>
    );
  };

  const CircleWithLine = ({ className }: { className?: string }) => {
    const id = useId();
    return (
      <div
        className={cn("flex flex-col items-center justify-center", className)}
      >
        <div
          className={cn(
            `h-3 w-3 rounded-full border border-neutral-200 bg-neutral-100 shadow-[2px_4px_16px_0px_rgba(248,248,248,0.06)_inset] dark:border-[rgba(255,255,255,0.2)] dark:bg-[rgba(248,248,248,0.02)]`
          )}
        />
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="2"
          height="265"
          viewBox="0 0 2 265"
          fill="none"
          className="h-full dark:text-white"
        >
          <path
            d="M1 265V1"
            stroke={`url(#${id})`}
            strokeOpacity="0.1"
            strokeWidth="1.5"
            strokeLinecap="round"
          />
          <defs>
            <linearGradient
              id={id}
              x1="1.5"
              y1="1"
              x2="1.5"
              y2="265"
              gradientUnits="userSpaceOnUse"
            >
              <stop stopColor="currentColor" stopOpacity="0.05" />
              <stop
                offset="0.530519"
                stopColor="currentColor"
                stopOpacity="0.5"
              />
              <stop offset="1" stopColor="currentColor" stopOpacity="0.05" />
            </linearGradient>
          </defs>
        </svg>
      </div>
    );
  };
  return (
    <div className="group/bento relative flex h-40 flex-col overflow-hidden px-2 py-8">
      <div className="absolute inset-0 flex h-full w-full flex-shrink-0 flex-row justify-center gap-4">
        <CircleWithLine />
        <CircleWithLine />
        <CircleWithLine />
        <CircleWithLine />
        <CircleWithLine />
        <CircleWithLine />
        <CircleWithLine />
        <CircleWithLine />
        <CircleWithLine />
        <CircleWithLine />
        <CircleWithLine />
        <CircleWithLine />
        <CircleWithLine />
      </div>
      <Container className="ml-20 mt-2">Application Submitted</Container>
      <Container className="ml-40 mt-4 transition duration-200 group-hover/bento:scale-[1.02] group-hover/bento:border-neutral-300 dark:group-hover/bento:border-neutral-500">
        Interview started
      </Container>
      <Cursor
        className="left-20 top-20 group-hover/bento:left-32"
        textClassName="group-hover/bento:text-neutral-500"
      />
    </div>
  );
};

export const SkeletonThree = () => {
  const Beam = ({
    className,
    delay,
    duration,
    width = 600,
    ...svgProps
  }: {
    className?: string;
    delay?: number;
    duration?: number;

    width?: number;
  } & React.ComponentProps<typeof motion.svg>) => {
    const id = useId();

    return (
      <motion.svg
        width={width ?? "600"}
        height="1"
        viewBox={`0 0 ${width ?? "600"} 1`}
        fill="none"
        xmlns="http://www.w3.org/2000/svg"
        className={cn("absolute inset-x-0 w-full", className)}
        {...svgProps}
      >
        <motion.path
          d={`M0 0.5H${width ?? "600"}`}
          stroke={`url(#svgGradient-${id})`}
        />

        <defs>
          <motion.linearGradient
            id={`svgGradient-${id}`}
            gradientUnits="userSpaceOnUse"
            initial={{ x1: "0%", x2: "-5%", y1: 0, y2: 0 }}
            animate={{ x1: "110%", x2: "105%", y1: 0, y2: 0 }}
            transition={{
              duration: duration ?? 2,
              ease: "linear",
              repeat: Infinity,
              delay: delay,
              repeatDelay: Math.random() * (2 - 1) + 1,
            }}
          >
            <stop stopColor="#2EB9DF" stopOpacity="0" />
            <stop stopColor="#3b82f6" />
            <stop offset="1" stopColor="#3b82f6" stopOpacity="0" />
          </motion.linearGradient>
        </defs>
      </motion.svg>
    );
  };

  const firstImageTextVariants = {
    initial: { y: 20, opacity: 0 },
    animate: { y: -30, opacity: 1 },
  };

  return (
    <motion.div className="relative pb-4">
      <Beam className="left-0 top-2 w-full" delay={0.5} />
      <Beam className="left-0 top-4 w-full" delay={0} />
      <Beam className="left-0 top-6 w-full" delay={1} />
      <Beam className="left-0 top-6 w-full" delay={3} />
      <Beam className="left-0 top-10 w-full" delay={2} />
      <Beam className="left-0 top-12 w-full" delay={3} />
      <Beam className="left-0 top-14 w-full" delay={0.2} />
      <Beam className="left-0 top-16 w-full" delay={1.2} />
      <Beam className="left-0 top-20 w-full" delay={2.3} />
      <div className="relative z-20 mx-auto grid max-w-[calc(100%-4rem)] grid-cols-2 justify-center gap-4">
        <div className="relative rounded-lg border border-neutral-200 bg-neutral-100 p-2 dark:border-neutral-700 dark:bg-neutral-800">
          <motion.div
            variants={firstImageTextVariants}
            className="absolute inset-x-0 bottom-0 mx-auto w-fit rounded-md bg-white px-2 py-1 text-xs text-black opacity-0 dark:bg-black dark:text-white"
          >
            Hello, My name...
          </motion.div>
          <Image
            src="https://assets.aceternity.com/pro/headshots/headshot-5.png"
            alt="avatar"
            width={100}
            height={100}
            className="h-full w-full rounded-lg object-cover object-top"
          />
        </div>
        <div className="relative rounded-lg border border-neutral-200 bg-neutral-100 p-2 dark:border-neutral-700 dark:bg-neutral-800">
          <motion.div
            variants={firstImageTextVariants}
            transition={{ delay: 0.5 }}
            className="absolute inset-x-0 bottom-0 mx-auto w-fit rounded-md bg-white px-2 py-1 text-xs text-black opacity-0 dark:bg-black dark:text-white"
          >
            lol, invert a <br /> binary tree
          </motion.div>
          <Image
            src="https://assets.aceternity.com/pro/headshots/headshot-6.png"
            alt="avatar"
            width={100}
            height={100}
            className="h-full w-full rounded-lg object-cover object-top"
          />
        </div>
      </div>
    </motion.div>
  );
};

export const SkeletonFour = ({}: {}) => {
  const [files, setFiles] = useState<File[]>([]);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleFileChange = (newFiles: File[]) => {
    setFiles((prevFiles) => [...prevFiles, ...newFiles]);
  };

  const handleClick = () => {
    fileInputRef.current?.click();
  };

  return (
    <div className="h-full w-full overflow-hidden">
      <motion.div
        onClick={handleClick}
        whileHover="animate"
        className="relative block h-full w-full cursor-pointer overflow-hidden rounded-none"
      >
        <input
          ref={fileInputRef}
          id="file-upload-handle"
          type="file"
          onChange={(e) => handleFileChange(Array.from(e.target.files || []))}
          className="hidden"
        />
        <div className="absolute inset-0 [mask-image:linear-gradient(to_top,white,transparent)]">
          <GridPattern />
        </div>
        <div className="flex h-full max-h-96 flex-col items-center justify-center overflow-hidden [mask-image:linear-gradient(to_top,transparent,white,transparent)]">
          <div className="relative mx-auto w-full max-w-xs">
            {files.length > 0 &&
              files.map((file, idx) => (
                <motion.div
                  key={"file" + idx}
                  layoutId={idx === 0 ? "file-upload" : "file-upload-" + idx}
                  className={cn(
                    "relative z-40 mx-auto mt-4 flex w-full flex-col items-start justify-start overflow-hidden rounded-md bg-white p-4 md:h-24 dark:bg-neutral-900",
                    "border-t border-neutral-100 shadow-md dark:border-neutral-800"
                  )}
                >
                  <div className="flex w-full items-center justify-between gap-4">
                    <motion.p
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      layout
                      className="max-w-xs truncate text-xs text-neutral-700 dark:text-neutral-300"
                    >
                      {file.name}
                    </motion.p>
                    <motion.p
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      layout
                      className="w-fit flex-shrink-0 rounded-lg px-2 py-1 text-xs text-neutral-600 shadow-md dark:bg-neutral-800 dark:text-white"
                    >
                      {(file.size / (1024 * 1024)).toFixed(2)} MB
                    </motion.p>
                  </div>

                  <div className="mt-2 flex w-full flex-col items-start justify-between text-sm text-neutral-600 md:flex-row md:items-center dark:text-neutral-400">
                    <motion.p
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      layout
                      className="rounded-md bg-gray-100 px-1 py-0.5 dark:bg-neutral-800"
                    >
                      {file.type}
                    </motion.p>

                    <motion.p
                      initial={{ opacity: 0 }}
                      animate={{ opacity: 1 }}
                      layout
                    >
                      modified{" "}
                      {new Date(file.lastModified).toLocaleDateString()}
                    </motion.p>
                  </div>
                </motion.div>
              ))}
            {!files.length && (
              <motion.div
                layoutId="file-upload"
                variants={mainVariant}
                transition={{ type: "spring", stiffness: 300, damping: 20 }}
                className={cn(
                  "relative z-40 mx-auto mt-4 flex h-32 w-full max-w-[8rem] items-center justify-center rounded-md bg-white group-hover/file:shadow-2xl dark:bg-neutral-900",
                  "shadow-[0px_10px_50px_rgba(0,0,0,0.1)]"
                )}
              >
                <IconUpload className="h-4 w-4 text-neutral-600 dark:text-neutral-300" />
              </motion.div>
            )}

            {!files.length && (
              <motion.div
                variants={secondaryVariant}
                className="absolute inset-0 z-30 mx-auto mt-4 flex h-32 w-full max-w-[8rem] items-center justify-center rounded-md border border-dashed border-sky-400 bg-transparent opacity-0"
              ></motion.div>
            )}
          </div>
        </div>
      </motion.div>
    </div>
  );
};
// Card structure
const CardSkeletonBody = ({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) => {
  return <div className={cn("", className)}>{children}</div>;
};

const CardContent = ({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) => {
  return <div className={cn("p-6", className)}>{children}</div>;
};

const CardTitle = ({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) => {
  return (
    <h3
      className={cn(
        "font-sans text-sm font-medium tracking-tight text-neutral-700 dark:text-neutral-100",
        className
      )}
    >
      {children}
    </h3>
  );
};
const CardDescription = ({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) => {
  return (
    <h3
      className={cn(
        "mt-2 max-w-xs font-sans text-sm font-normal tracking-tight text-neutral-500 dark:text-neutral-400",
        className
      )}
    >
      {children}
    </h3>
  );
};

const Card = ({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) => {
  return (
    <motion.div
      whileHover="animate"
      className={cn(
        "group isolate flex flex-col overflow-hidden rounded-2xl bg-white shadow-[0_1px_1px_rgba(0,0,0,0.05),0_4px_6px_rgba(34,42,53,0.04),0_24px_68px_rgba(47,48,55,0.05),0_2px_3px_rgba(0,0,0,0.04)] dark:bg-neutral-900",
        className
      )}
    >
      {children}
    </motion.div>
  );
};

// Variants

const mainVariant = {
  initial: { x: 0, y: 0 },
  animate: { x: 20, y: -20, opacity: 0.9 },
};

const secondaryVariant = { initial: { opacity: 0 }, animate: { opacity: 1 } };

// Utils

export function GridPattern() {
  const columns = 20;
  const rows = 5;
  return (
    <div className="flex flex-shrink-0 scale-110 flex-wrap items-center justify-center gap-x-px gap-y-px bg-gray-100 dark:bg-neutral-900">
      {Array.from({ length: rows }).map((_, row) =>
        Array.from({ length: columns }).map((_, col) => {
          const index = row * columns + col;
          return (
            <div
              key={`${col}-${row}-grid-pattern`}
              className={`flex h-10 w-10 flex-shrink-0 rounded-[2px] ${
                index % 2 === 0
                  ? "bg-gray-50 dark:bg-neutral-950"
                  : "bg-gray-50 shadow-[0px_0px_1px_3px_rgba(255,255,255,1)_inset] dark:bg-neutral-950 dark:shadow-[0px_0px_1px_3px_rgba(0,0,0,1)_inset]"
              }`}
            />
          );
        })
      )}
    </div>
  );
}
