"use client";
import Image from "next/image";
import Marquee from "react-fast-marquee";

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
    <div className="relative z-20 px-4 py-10 md:px-8 md:py-40">
      <h2 className="bg-gradient-to-b from-neutral-900 to-neutral-600 bg-clip-text text-center font-sans text-2xl font-bold text-transparent md:text-5xl dark:from-white dark:to-neutral-600">
        Brands love us
      </h2>
      <p className="mt-4 text-center font-sans text-base text-neutral-700 dark:text-neutral-300">
        Aceternity UI is loved by the best companies who are serious about what
        they do.
      </p>

      <div className="relative mx-auto mt-20 flex h-full w-full max-w-7xl flex-wrap justify-center gap-10 [mask-image:linear-gradient(to_right,transparent,black_20%,black_80%,transparent)]">
        <Marquee pauseOnHover direction="right">
          {logos.map((logo, idx) => (
            <Image
              key={logo.name + "logo-marquee" + idx}
              src={logo.src}
              alt={logo.name}
              width="100"
              height="100"
              className="mx-0 w-32 object-contain filter md:mx-10 md:w-40 dark:invert"
            />
          ))}
        </Marquee>
      </div>
      <div className="relative mx-auto mt-4 flex h-full w-full max-w-7xl flex-wrap justify-center gap-10 [mask-image:linear-gradient(to_right,transparent,black_20%,black_80%,transparent)] md:mt-20 md:gap-40">
        <Marquee pauseOnHover direction="left" speed={30}>
          {logos.map((logo, idx) => (
            <Image
              key={logo.name + "logo-marquee-second" + idx}
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
