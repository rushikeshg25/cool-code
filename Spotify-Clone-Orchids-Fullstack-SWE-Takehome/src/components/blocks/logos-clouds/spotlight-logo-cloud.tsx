"use client";
import Image from "next/image";
import React from "react";

export function SpotlightLogoCloud() {
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
      name: "Asteroid Kit 2",
      src: "https://assets.aceternity.com/pro/logos/asteroid-kit.png",
    },

    {
      name: "Host it 2",
      src: "https://assets.aceternity.com/pro/logos/hostit.png",
    },
    {
      name: "Aceternity UI 2",
      src: "https://assets.aceternity.com/pro/logos/aceternity-ui.png",
    },
    {
      name: "Gamity 2",
      src: "https://assets.aceternity.com/pro/logos/gamity.png",
    },
  ];

  return (
    <div className="relative w-full overflow-hidden py-40">
      <AmbientColor />
      <h2 className="bg-gradient-to-b from-neutral-900 to-neutral-600 bg-clip-text pb-4 text-center font-sans text-2xl font-bold text-transparent md:text-5xl dark:from-white dark:to-neutral-600">
        Brands with a spotlight
      </h2>
      <p className="text-neutral-7000 mb-10 mt-4 text-center font-sans text-base dark:text-neutral-300">
        Brands who funded us deserve more than a spotlight. Check out what they
        are saying.
      </p>
      <div className="relative mx-auto grid w-full max-w-3xl grid-cols-4 gap-10">
        {logos.map((logo, idx) => (
          <div
            key={logo.src + idx + "logo-spotlight"}
            className="flex items-center justify-center"
          >
            <Image
              src={logo.src}
              alt={logo.name}
              width={100}
              height={100}
              className="w-full select-none object-contain dark:invert dark:filter"
              draggable={false}
            />
          </div>
        ))}
      </div>
    </div>
  );
}

export const AmbientColor = () => {
  return (
    <div className="pointer-events-none absolute left-40 top-0 z-40 h-screen w-screen">
      <div
        style={{
          transform: "translateY(-350px) rotate(-45deg)",
          width: "560px",
          height: "1380px",
          background:
            "radial-gradient(68.54% 68.72% at 55.02% 31.46%, hsla(240, 100%, 85%, .2) 0, hsla(240, 100%, 55%, .1) 50%, hsla(240, 100%, 45%, .05) 80%)",
          filter: "blur(20px)",
          borderRadius: "50%",
        }}
        className="absolute left-0 top-0"
      />

      <div
        style={{
          transform: "rotate(-45deg) translate(5%, -50%)",
          transformOrigin: "top left",
          width: "240px",
          height: "1380px",
          background:
            "radial-gradient(50% 50% at 50% 50%, hsla(240, 100%, 85%, .15) 0, hsla(240, 100%, 45%, .1) 80%, transparent 100%)",
          filter: "blur(20px)",
          borderRadius: "50%",
        }}
        className="absolute left-0 top-0"
      />

      <div
        style={{
          position: "absolute",
          borderRadius: "50%",
          transform: "rotate(-45deg) translate(-180%, -70%)",
          transformOrigin: "top left",
          top: 0,
          left: 0,
          width: "240px",
          height: "1380px",
          background:
            "radial-gradient(50% 50% at 50% 50%, hsla(240, 100%, 85%, .1) 0, hsla(240, 100%, 45%, .05) 80%, transparent 100%)",
          filter: "blur(20px)",
        }}
        className="absolute left-0 top-0"
      />
    </div>
  );
};
