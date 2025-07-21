"use client";
import React from "react";
import { cn } from "@/lib/utils";
import Image from "next/image";
import Marquee from "react-fast-marquee";

export function TestimonialsMarqueeGrid() {
  return (
    <div className="relative w-full max-w-7xl mx-auto px-4 md:px-8 pt-20 overflow-hidden h-full ">
      <div className="pb-20">
        <h1 className="pt-4 font-bold text-black text-lg md:text-2xl dark:text-white">
          Used by builders around the world
        </h1>
        <p className="text-base mt-4 text-neutral-600 dark:text-neutral-200">
          Everyone uses our product, except for the people who don&apos;t use
          it.
        </p>
      </div>

      <div className=" relative">
        <div className="h-full overflow-hidden w-full bg-charcoal">
          <TestimonialsGrid />
        </div>
      </div>

      <div className="absolute bottom-0 inset-x-0 h-40 w-full bg-gradient-to-t from-charcoal to-transparent"></div>
    </div>
  );
}

export const TestimonialsGrid = () => {
  const first = testimonials.slice(0, 6);
  const second = testimonials.slice(6, 12);

  return (
    <div className="relative [mask-image:linear-gradient(to_right,transparent_0%,white_10%,white_90%,transparent_100%)]">
      <Marquee direction="right" pauseOnHover speed={50}>
        {first.map((testimonial, index) => (
          <Card key={`testimonial-${testimonial.src}-${index}`}>
            <Quote>{testimonial.quote}</Quote>
            <div className="flex gap-2 items-center mt-8">
              <Image
                src={testimonial.src}
                alt="Manu Arora"
                width={40}
                height={40}
                className="rounded-full"
              />
              <div className="flex flex-col">
                <QuoteDescription className="text-neutral-600 dark:text-neutral-300">
                  {testimonial.name}
                </QuoteDescription>
                <QuoteDescription className="text-neutral-400">
                  {testimonial.designation}
                </QuoteDescription>
              </div>
            </div>
          </Card>
        ))}
      </Marquee>
      <Marquee className="mt-10" direction="right" pauseOnHover speed={70}>
        {second.map((testimonial, index) => (
          <Card key={`testimonial-${testimonial.src}-${index}`}>
            <Quote>{testimonial.quote}</Quote>
            <div className="flex gap-2 items-center mt-8">
              <Image
                src={testimonial.src}
                alt="Manu Arora"
                width={40}
                height={40}
                className="rounded-full"
              />
              <div className="flex flex-col">
                <QuoteDescription className="text-neutral-300">
                  {testimonial.name}
                </QuoteDescription>
                <QuoteDescription className="text-neutral-400">
                  {testimonial.designation}
                </QuoteDescription>
              </div>
            </div>
          </Card>
        ))}
      </Marquee>
    </div>
  );
};
export const Card = ({
  className,
  children,
}: {
  className?: string;
  children: React.ReactNode;
}) => {
  return (
    <div
      className={cn(
        "p-4 md:p-8 rounded-xl min-h-[230px] h-full max-w-md md:max-w-lg mx-4 bg-gray-50  dark:bg-neutral-900 shadow-[2px_4px_16px_0px_rgba(248,248,248,0.06)_inset] group",
        className
      )}
    >
      {children}
    </div>
  );
};

export const Quote = ({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) => {
  return (
    <h3
      className={cn(
        "text-sm md:text-base font-semibold dark:text-white text-black py-2",
        className
      )}
    >
      {children}
    </h3>
  );
};

export const QuoteDescription = ({
  children,
  className,
}: {
  children: React.ReactNode;
  className?: string;
}) => {
  return (
    <p
      className={cn(
        "text-xs md:text-sm font-normal dark:text-neutral-400 text-neutral-600 max-w-sm",
        className
      )}
    >
      {children}
    </p>
  );
};

interface Testimonial {
  src: string;
  quote: string;
  name: string;
  designation?: string;
}

export const testimonials: Testimonial[] = [
  {
    name: "Manu Arora",
    quote:
      "What a fantastic AI Proactiv AI is, I just love it. It has completely transformed the way I approach problems and develop solutions.",
    src: "https://i.pravatar.cc/150?img=1",
    designation: "Tech Innovator & Entrepreneur",
  },
  {
    name: "Tyler Durden",
    quote:
      "I made a soap with the help of AI, it was so easy to use. I'm so glad this happened because it revolutionized my entire business model and production process.",
    src: "https://i.pravatar.cc/150?img=2",
    designation: "Creative Director & Business Owner",
  },
  {
    name: "Alice Johnson",
    quote:
      "This AI has transformed the way I work! It's like having a brilliant assistant who knows exactly what I need before I even ask.",
    src: "https://i.pravatar.cc/150?img=3",
    designation: "Senior Software Engineer",
  },
  {
    name: "Bob Smith",
    quote:
      "Absolutely revolutionary, a game-changer for our industry. It has streamlined our processes and enhanced our productivity dramatically.",
    src: "https://i.pravatar.cc/150?img=4",
    designation: "Industry Analyst",
  },
  {
    name: "Cathy Lee",
    quote:
      "I can't imagine going back to how things were before this AI. It has not only improved my work efficiency but also my daily life.",
    src: "https://i.pravatar.cc/150?img=5",
    designation: "Product Manager",
  },
  {
    name: "David Wright",
    quote:
      "It's like having a superpower! This AI tool has given us the ability to do things we never thought were possible in our field.",
    src: "https://i.pravatar.cc/150?img=6",
    designation: "Research Scientist",
  },
  {
    name: "Eva Green",
    quote:
      "The efficiency it brings is unmatched. It's a vital tool that has helped us cut costs and improve our end product significantly.",
    src: "https://i.pravatar.cc/150?img=7",
    designation: "Operations Director",
  },
  {
    name: "Frank Moore",
    quote:
      "A robust solution that fits perfectly into our workflow. It has enhanced our team's capabilities and allowed us to tackle more complex projects.",
    src: "https://i.pravatar.cc/150?img=8",
    designation: "Project Manager",
  },
  {
    name: "Grace Hall",
    quote:
      "It's incredibly intuitive and easy to use. Even those without technical expertise can leverage its power to improve their workflows.",
    src: "https://i.pravatar.cc/150?img=9",
    designation: "Marketing Specialist",
  },
  {
    name: "Henry Ford",
    quote:
      "It has saved us countless hours. Highly recommended for anyone looking to enhance their efficiency and productivity.",
    src: "https://i.pravatar.cc/150?img=10",
    designation: "Operations Analyst",
  },
  {
    name: "Ivy Wilson",
    quote:
      "A must-have tool for any professional. It's revolutionized the way we approach problem-solving and decision-making.",
    src: "https://i.pravatar.cc/150?img=11",
    designation: "Business Consultant",
  },
  {
    name: "Jack Brown",
    quote:
      "The results are always impressive. This AI has helped us to not only meet but exceed our performance targets.",
    src: "https://i.pravatar.cc/150?img=12",
    designation: "Performance Manager",
  },
  {
    name: "Kathy Adams",
    quote:
      "It helps us achieve what was once thought impossible. The AI's capabilities are groundbreaking and have opened new avenues for us.",
    src: "https://i.pravatar.cc/150?img=13",
    designation: "Innovation Lead",
  },
  {
    name: "Leo Carter",
    quote:
      "Transformative technology with real impact. It has streamlined our operations and brought unprecedented efficiency to our processes.",
    src: "https://i.pravatar.cc/150?img=14",
    designation: "Technology Strategist",
  },
];
