"use client";
import { IconMenu2, IconX } from "@tabler/icons-react";
import { motion, AnimatePresence } from "motion/react";
import Link from "next/link";
import React, { useState } from "react";
import Image from "next/image";
import Marquee from "react-fast-marquee";
import { cn } from "@/lib/utils";

export function HeroSectionWithImagesGrid() {
  return (
    <div className="relative w-full overflow-hidden bg-gray-50 dark:bg-neutral-950">
      <Navbar />
      <div className="relative flex flex-col items-center justify-center overflow-hidden px-8 pb-4 md:px-8">
        <div className="relative mt-20 flex flex-col items-center justify-center">
          <FeaturedImages />
          <h1 className="mb-8relative mx-auto mt-4 max-w-6xl text-center text-3xl font-bold tracking-tight text-zinc-700 md:text-4xl lg:text-7xl dark:text-white">
            Your best in class{" "}
            <span className="relative z-10 bg-gradient-to-b from-indigo-700 to-indigo-600 bg-clip-text text-transparent">
              design and development studio
            </span>{" "}
            <span>
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="inline-block h-14 w-14 stroke-indigo-500 stroke-[1px]"
                width="24"
                height="24"
                viewBox="0 0 24 24"
                strokeWidth="2"
                stroke="currentColor"
                fill="none"
                strokeLinecap="round"
                strokeLinejoin="round"
              >
                <motion.path stroke="none" d="M0 0h24v24H0z" fill="none" />
                <motion.path
                  initial={{ pathLength: 0, fill: "#a5b4fc", opacity: 0 }}
                  animate={{ pathLength: 1, fill: "#a5b4fc", opacity: 1 }}
                  transition={{
                    duration: 1.5,
                    repeat: Infinity,
                    repeatType: "reverse",
                    ease: "linear",
                    repeatDelay: 0.5,
                  }}
                  d="M13 3l0 7l6 0l-8 11l0 -7l-6 0l8 -11"
                />
              </svg>
            </span>
          </h1>
          <h2 className="font-regular relative mx-auto mb-8 mt-8 max-w-xl text-center text-base tracking-wide text-zinc-500 antialiased md:text-xl dark:text-zinc-200">
            We provide the best in class design and development services for
            teams that ship with the speed of light.
          </h2>
        </div>
        <div className="group relative z-10 mb-10">
          <button className="rounded-lg bg-black px-8 py-2 font-medium text-white shadow-[0px_-2px_0px_0px_rgba(255,255,255,0.4)_inset] dark:bg-white dark:text-black">
            Book a call
          </button>
        </div>
        <LogoCloudMarquee />
      </div>
      <ImagesGrid />
    </div>
  );
}

export const ImagesGrid = () => {
  const images = [
    {
      src: "https://assets.aceternity.com/pro/hero-example-3.jpg",
      className: "translate-y-10",
    },
    {
      src: "https://assets.aceternity.com/pro/hero-example-1.jpg",
      className: "translate-y-20",
    },

    {
      src: "https://assets.aceternity.com/pro/hero-example-2.jpg",
      className: "translate-y-4",
    },
    {
      src: "https://assets.aceternity.com/pro/hero-example-4.jpg",
      className: "translate-y-10",
    },
    {
      src: "https://assets.aceternity.com/pro/hero-example-5.jpg",
      className: "translate-y-20",
    },
  ];
  return (
    <div className="relative mt-10 h-[20rem] w-full overflow-hidden border-b border-neutral-200 md:h-[30rem] dark:border-neutral-800">
      <div className="absolute inset-0 flex h-full w-full flex-shrink-0 justify-center gap-5">
        {images.map((image) => (
          <div
            className={cn(
              "relative mt-0 rounded-lg border border-neutral-200 bg-gray-100 p-2 dark:border-neutral-900 dark:bg-neutral-800",
              image.className
            )}
            key={image.src}
          >
            <Image
              src={image.src}
              alt={image.src}
              width="500"
              height="500"
              className="h-full min-w-[15rem] flex-shrink-0 rounded-lg object-cover object-top"
            />
          </div>
        ))}
      </div>
    </div>
  );
};

export const FeaturedImages = ({
  className,
  containerClassName,
}: {
  textClassName?: string;
  className?: string;
  showStars?: boolean;
  containerClassName?: string;
}) => {
  const images = [
    {
      name: "John Doe",
      src: "https://images.unsplash.com/photo-1599566150163-29194dcaad36?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=3387&q=80",
    },
    {
      name: "Robert Johnson",
      src: "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxzZWFyY2h8Mnx8YXZhdGFyfGVufDB8fDB8fHww&auto=format&fit=crop&w=800&q=60",
    },
    {
      name: "Jane Smith",
      src: "https://images.unsplash.com/photo-1580489944761-15a19d654956?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxzZWFyY2h8NXx8YXZhdGFyfGVufDB8fDB8fHww&auto=format&fit=crop&w=800&q=60",
    },
    {
      name: "Emily Davis",
      src: "https://images.unsplash.com/photo-1438761681033-6461ffad8d80?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxzZWFyY2h8MTB8fGF2YXRhcnxlbnwwfHwwfHx8MA%3D%3D&auto=format&fit=crop&w=800&q=60",
    },
    {
      name: "Tyler Durden",
      src: "https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=3540&q=80",
    },
    {
      name: "Dora",
      src: "https://images.unsplash.com/photo-1544725176-7c40e5a71c5e?ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D&auto=format&fit=crop&w=3534&q=80",
    },
  ];
  return (
    <div className={cn("flex flex-col items-center", containerClassName)}>
      <div
        className={cn(
          "mb-2 flex flex-col items-center justify-center sm:flex-row",
          className
        )}
      >
        <div className="mb-4 flex flex-row items-center sm:mb-0">
          {images.map((image, idx) => (
            <div className="group relative -mr-4" key={image.name}>
              <div>
                <motion.div
                  whileHover={{ scale: 1.05, zIndex: 30 }}
                  transition={{ duration: 0.2 }}
                  className="relative overflow-hidden rounded-full border-2 border-neutral-200"
                >
                  <Image
                    height={100}
                    width={100}
                    src={image.src}
                    alt={image.name}
                    className="h-8 w-8 object-cover object-top md:h-14 md:w-14"
                  />
                </motion.div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export function LogoCloudMarquee() {
  const logos = [
    {
      name: "Aceternity UI",
      src: "https://assets.aceternity.com/pro/logos/aceternity-ui.png",
    },
    {
      name: "Gamity",
      src: "https://assets.aceternity.com/pro/logos/gamity.png",
    },
    {
      name: "Host it",
      src: "https://assets.aceternity.com/pro/logos/hostit.png",
    },
    {
      name: "Asteroid Kit",
      src: "https://assets.aceternity.com/pro/logos/asteroid-kit.png",
    },
    {
      name: "Aceternity UI 2",
      src: "https://assets.aceternity.com/pro/logos/aceternity-ui.png",
    },
    {
      name: "Gamity 2",
      src: "https://assets.aceternity.com/pro/logos/gamity.png",
    },
    {
      name: "Host it 2",
      src: "https://assets.aceternity.com/pro/logos/hostit.png",
    },
    {
      name: "Asteroid Kit 2",
      src: "https://assets.aceternity.com/pro/logos/asteroid-kit.png",
    },
  ];

  return (
    <div className="relative">
      <p className="mt-4 text-center font-sans text-base text-neutral-700 dark:text-neutral-300">
        Trusted by famous brands
      </p>

      <div className="relative mx-auto mt-4 flex h-20 w-full max-w-4xl flex-wrap justify-center gap-10 [mask-image:linear-gradient(to_right,transparent,black_20%,black_80%,transparent)] md:mt-2 md:gap-40">
        <Marquee pauseOnHover direction="left" speed={30}>
          {logos.map((logo, idx) => (
            <Image
              key={logo.name + "second"}
              src={logo.src}
              alt={logo.name}
              width="100"
              height="100"
              className="mx-0 w-32 object-contain filter md:mx-10 md:w-40 dark:invert"
            />
          ))}
        </Marquee>
      </div>
    </div>
  );
}

const Navbar = () => {
  const navItems = [
    { name: "Work", link: "#" },
    { name: "Services", link: "#" },
    { name: "Pricing", link: "#" },
    { name: "Contact", link: "#" },
  ];
  return (
    <div className="relative z-[60] mx-auto flex w-full max-w-7xl flex-row items-center justify-between px-8 py-8">
      <Logo />
      <div className="hidden flex-1 flex-row items-center justify-center space-x-8 text-sm font-medium text-zinc-600 transition duration-200 hover:text-zinc-800 lg:flex lg:space-x-14">
        <DesktopNav navItems={navItems} />
      </div>
      <button className="hidden rounded-lg bg-black px-8 py-2 font-medium text-white shadow-[0px_-2px_0px_0px_rgba(255,255,255,0.4)_inset] md:block dark:bg-white dark:text-black">
        Book a call
      </button>

      <div className="flex lg:hidden">
        <MobileNav navItems={navItems} />
      </div>
    </div>
  );
};

const DesktopNav = ({ navItems }: any) => {
  return (
    <>
      {navItems.map((navItem: any, idx: number) => (
        <Link
          className="text-neutral-600 dark:text-neutral-300"
          key={`link=${idx}`}
          href={navItem.link}
        >
          <span>{navItem.name}</span>
        </Link>
      ))}
    </>
  );
};

const MobileNav = ({ navItems }: any) => {
  const [open, setOpen] = useState(false);

  return (
    <>
      <IconMenu2 onClick={() => setOpen(!open)} />
      <AnimatePresence>
        {open && (
          <motion.div className="fixed inset-0 z-50 flex flex-col items-center justify-center space-y-10 bg-white text-xl font-bold text-zinc-600 transition duration-200 hover:text-zinc-800">
            <IconX
              className="absolute right-8 top-8 h-5 w-5"
              onClick={() => setOpen(!open)}
            />
            {navItems.map((navItem: any, idx: number) => (
              <Link
                key={`link=${idx}`}
                href={navItem.link}
                className="relative text-neutral-600 dark:text-neutral-300"
              >
                <motion.span className="block">{navItem.name} </motion.span>
              </Link>
            ))}
            <button className="rounded-lg bg-black px-8 py-2 font-medium text-white shadow-[0px_-2px_0px_0px_rgba(255,255,255,0.4)_inset] dark:bg-white dark:text-black">
              Book a call
            </button>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  );
};

const Logo = () => {
  return (
    <Link
      href="/"
      className="relative z-20 mr-4 flex items-center space-x-2 px-2 py-1 text-sm font-normal text-black"
    >
      <div className="h-5 w-6 rounded-bl-sm rounded-br-lg rounded-tl-lg rounded-tr-sm bg-black dark:bg-white" />
      <span className="font-medium text-black dark:text-white">DevStudio</span>
    </Link>
  );
};
