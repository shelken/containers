import { useRef } from 'react';
import { useGSAP } from '@gsap/react';
import gsap from 'gsap';
import { ScrollTrigger } from 'gsap/ScrollTrigger';

const navItems = ['Docs', 'Changelog', 'Blog', 'Guide'];

const bentoCards = [
  {
    title: 'Tiny surface, cinematic finish',
    body: 'A single static page with glass navigation, editorial rhythm, and enough motion to feel alive without becoming heavy.',
    className: 'lg:col-span-7',
    image: 'https://picsum.photos/seed/hello-polished-glass/1280/900',
  },
  {
    title: 'Fast by default',
    body: 'Vite output, immutable assets, and unprivileged Nginx keep the service small, predictable, and easy to ship.',
    className: 'lg:col-span-5',
    image: 'https://picsum.photos/seed/hello-fast-light/1280/900',
  },
  {
    title: 'Copy, edit, launch',
    body: 'The template flow stays obvious: rename, replace the page, validate locally, then let CI publish the image.',
    className: 'lg:col-span-4',
    image: 'https://picsum.photos/seed/hello-launch-room/1280/900',
  },
  {
    title: 'Micro-interactions everywhere',
    body: 'Pills, cards, and images lift, glow, and sweep with restrained hover physics inspired by premium docs navigation.',
    className: 'lg:col-span-4',
    image: 'https://picsum.photos/seed/hello-motion-soft/1280/900',
  },
  {
    title: 'GitOps ready',
    body: 'A finished image maps cleanly into simple-service deployment resources in home-ops.',
    className: 'lg:col-span-4',
    image: 'https://picsum.photos/seed/hello-gitops-map/1280/900',
  },
];

const accordionItems = [
  {
    title: 'Glass nav',
    body: 'Layered shadows, high contrast target states, and animated sheen return on hover.',
    image: 'https://picsum.photos/seed/hello-nav-glass/900/1200',
  },
  {
    title: 'Editorial scale',
    body: 'Wide headlines avoid cramped wrapping and keep the first impression calm.',
    image: 'https://picsum.photos/seed/hello-editorial-scale/900/1200',
  },
  {
    title: 'Static runtime',
    body: 'No backend path, no database, no secrets. The page remains easy to reason about.',
    image: 'https://picsum.photos/seed/hello-static-runtime/900/1200',
  },
];

const marqueeWords = ['Hello World', 'Static Service', 'Glass UI', 'Fast Preview', 'GitOps Flow'];

function ArrowIcon() {
  return (
    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" className="-translate-x-1.5 opacity-0 transition-all duration-300 ease-out group-hover/nav:translate-x-0 group-hover/nav:opacity-100">
      <path d="M7 17 17 7" />
      <path d="M8 7h9v9" />
    </svg>
  );
}

function NavPill({ label }: { label: string }) {
  return (
    <a href={`#${label.toLowerCase()}`} className="group/nav relative hidden h-9 items-center overflow-hidden rounded-full px-3.5 text-sm font-medium tracking-tight text-neutral-700 outline-none transition-[color,transform,box-shadow] duration-300 ease-[cubic-bezier(0.34,1.56,0.64,1)] hover:-translate-y-px hover:text-neutral-950 hover:shadow-[0_6px_16px_-8px_rgba(15,15,15,0.22)] focus-visible:ring-2 focus-visible:ring-neutral-400/80 focus-visible:ring-offset-2 focus-visible:ring-offset-[#f5f4f2] active:translate-y-0 active:scale-[0.98] sm:inline-flex">
      <span aria-hidden="true" className="sheen-sweep pointer-events-none absolute inset-0 bg-gradient-to-r from-transparent via-white/50 to-transparent" />
      <span aria-hidden="true" className="absolute inset-0 origin-left scale-x-0 rounded-full bg-neutral-100/95 transition-transform duration-500 ease-[cubic-bezier(0.16,1,0.3,1)] group-hover/nav:scale-x-100" />
      <span className="relative z-10 inline-flex items-center gap-1">
        {label}
        <ArrowIcon />
      </span>
    </a>
  );
}

export default function App() {
  const pageRef = useRef<HTMLElement>(null);

  useGSAP(
    () => {
      gsap.registerPlugin(ScrollTrigger);

      gsap.from('[data-hero-line]', {
        y: 42,
        opacity: 0,
        duration: 1.1,
        stagger: 0.12,
        ease: 'power3.out',
      });

      gsap.utils.toArray<HTMLElement>('[data-scale-image]').forEach((image) => {
        gsap.fromTo(
          image,
          { scale: 0.84, opacity: 0.58, filter: 'brightness(0.72) contrast(1.2)' },
          {
            scale: 1,
            opacity: 1,
            filter: 'brightness(1) contrast(1.08)',
            ease: 'none',
            scrollTrigger: {
              trigger: image,
              start: 'top 85%',
              end: 'bottom 25%',
              scrub: true,
            },
          },
        );
      });

      gsap.utils.toArray<HTMLElement>('[data-reveal-word]').forEach((word, index) => {
        gsap.fromTo(
          word,
          { opacity: 0.14, y: 10 },
          {
            opacity: 1,
            y: 0,
            ease: 'none',
            scrollTrigger: {
              trigger: '[data-reveal-copy]',
              start: `top+=${index * 8} 78%`,
              end: `top+=${index * 8 + 120} 42%`,
              scrub: true,
            },
          },
        );
      });
    },
    { scope: pageRef },
  );

  const revealWords = 'Hello World turns a tiny static container into a polished product surface with motion, contrast, and deployment discipline.'.split(' ');

  return (
    <main ref={pageRef} className="w-full max-w-full overflow-x-hidden bg-[#f5f4f2] text-neutral-950">
      <div className="pointer-events-none fixed inset-0 z-0 grain-overlay opacity-70" />
      <div className="pointer-events-none fixed left-1/2 top-[-12rem] z-0 h-[38rem] w-[62rem] -translate-x-1/2 rounded-full bg-[radial-gradient(circle,rgba(255,255,255,0.9),rgba(210,196,174,0.32)_42%,transparent_68%)] blur-3xl" />

      <header className="fixed left-0 right-0 top-5 z-50 flex justify-center px-4">
        <nav aria-label="Primary" className="pointer-events-auto inline-flex max-w-[min(100vw-2rem,52rem)] shrink-0 flex-wrap items-center justify-end gap-1 rounded-full border border-neutral-200/80 bg-white/80 px-1.5 py-1.5 shadow-[0_6px_18px_-6px_rgba(15,15,15,0.18),0_1px_2px_rgba(0,0,0,0.04)] backdrop-blur-md transition-[box-shadow,border-color] duration-300 ease-out hover:border-neutral-300/90 hover:shadow-[0_10px_26px_-10px_rgba(15,15,15,0.22),0_1px_2px_rgba(0,0,0,0.04)] sm:gap-2 sm:px-2 sm:max-w-none sm:flex-nowrap">
          {navItems.map((item) => (
            <NavPill key={item} label={item} />
          ))}
          <a href="https://github.com/shelken/containers" target="_blank" rel="noreferrer" className="group relative inline-flex h-9 items-center gap-1.5 overflow-hidden rounded-full bg-[#0a0a0a] px-3 text-[13px] font-semibold leading-[1.2] text-white shadow-[inset_0_1px_0_rgba(255,255,255,0.08)] ring-1 ring-black/20 transition-[background-color,box-shadow,transform] duration-500 ease-[cubic-bezier(0.16,1,0.3,1)] hover:bg-[#141414] hover:shadow-[0_14px_32px_-10px_rgba(0,0,0,0.55),inset_0_1px_0_rgba(255,255,255,0.12)] active:scale-[0.97]">
            <span aria-hidden="true" className="sheen-sweep pointer-events-none absolute inset-0 bg-gradient-to-r from-transparent via-white/25 to-transparent" />
            <span className="relative font-mono text-[12px] tracking-tight transition-transform duration-300 ease-out group-hover:-translate-y-px group-hover:scale-[1.03]">Source</span>
          </a>
          <button type="button" aria-label="Open menu" className="group ml-1 inline-flex h-8 w-8 items-center justify-center rounded-full border border-neutral-200 bg-white text-neutral-900 transition-[background-color,transform,box-shadow] duration-200 ease-out hover:bg-neutral-50 hover:shadow-[0_8px_18px_-12px_rgba(15,15,15,0.2)] active:scale-95 sm:hidden">
            <span className="relative block h-3 w-4">
              <span aria-hidden="true" className="absolute left-0 right-0 top-0 h-[1.6px] rounded-full bg-neutral-900 transition-transform duration-300 ease-[cubic-bezier(0.34,1.56,0.64,1)] group-hover:-translate-y-px" />
              <span aria-hidden="true" className="absolute left-0 right-0 top-1/2 h-[1.6px] -translate-y-1/2 rounded-full bg-neutral-900 transition-transform duration-200 ease-out group-hover:scale-x-[0.88]" />
              <span aria-hidden="true" className="absolute bottom-0 left-0 right-0 h-[1.6px] rounded-full bg-neutral-900 transition-transform duration-300 ease-[cubic-bezier(0.34,1.56,0.64,1)] group-hover:translate-y-px" />
            </span>
          </button>
        </nav>
      </header>

      <section className="relative z-10 flex min-h-screen items-center justify-center px-5 pb-24 pt-36 md:px-10">
        <div className="mx-auto w-full max-w-7xl text-center">
          <h1 className="mx-auto max-w-6xl text-[clamp(3.2rem,7vw,7.2rem)] font-semibold leading-[0.88] tracking-[-0.08em] text-neutral-950">
            <span data-hero-line className="block">Hello World,</span>
            <span data-hero-line className="block">but make it expensive</span>
          </h1>
          <p data-hero-line className="mx-auto mt-8 max-w-3xl text-balance text-lg leading-8 text-neutral-600 md:text-xl">
            A fancy static landing page that treats a tiny service like a polished launch moment: glass navigation, wide type, dense cards, and motion with restraint.
          </p>
          <div data-hero-line className="mt-10 flex flex-col items-center justify-center gap-3 sm:flex-row">
            <a href="#launch" className="group relative inline-flex h-12 items-center justify-center overflow-hidden rounded-full bg-neutral-950 px-6 text-sm font-semibold text-white shadow-[0_18px_38px_-16px_rgba(0,0,0,0.65)] transition-transform duration-300 ease-[cubic-bezier(0.34,1.56,0.64,1)] hover:-translate-y-0.5 active:scale-[0.98]">
              <span aria-hidden="true" className="sheen-sweep pointer-events-none absolute inset-0 bg-gradient-to-r from-transparent via-white/25 to-transparent" />
              <span className="relative">Launch locally</span>
            </a>
            <a href="#craft" className="inline-flex h-12 items-center justify-center rounded-full border border-neutral-200 bg-white/80 px-6 text-sm font-semibold text-neutral-950 shadow-[0_12px_30px_-18px_rgba(0,0,0,0.4)] backdrop-blur-md transition-[transform,border-color,box-shadow] duration-300 ease-out hover:-translate-y-0.5 hover:border-neutral-300 hover:shadow-[0_18px_34px_-22px_rgba(0,0,0,0.45)] active:scale-[0.98]">
              See the craft
            </a>
          </div>
          <div data-scale-image className="mx-auto mt-16 h-[18rem] max-w-6xl overflow-hidden rounded-[2rem] border border-white/70 bg-neutral-900 shadow-[0_40px_120px_-48px_rgba(0,0,0,0.72)] md:h-[30rem]">
            <img src="https://picsum.photos/seed/hello-cinematic-console/1920/1080" alt="Abstract studio light over a polished surface" className="h-full w-full object-cover opacity-90 mix-blend-luminosity contrast-125 transition-transform duration-700 ease-out hover:scale-105" />
          </div>
        </div>
      </section>

      <section id="craft" className="relative z-10 px-5 py-32 md:px-10 md:py-48">
        <div className="mx-auto max-w-7xl">
          <div className="mb-14 flex flex-col justify-between gap-8 md:flex-row md:items-end">
            <h2 className="max-w-4xl text-[clamp(2.6rem,5vw,5.2rem)] font-semibold leading-[0.92] tracking-[-0.07em]">
              Built like a product page,
              <span className="mx-3 inline-block h-12 w-28 overflow-hidden rounded-full align-middle md:h-16 md:w-40">
                <img src="https://picsum.photos/seed/hello-inline-typography/640/320" alt="Soft abstract texture" className="h-full w-full object-cover grayscale contrast-125" />
              </span>
              shipped like a tiny service.
            </h2>
            <p className="max-w-sm text-base leading-7 text-neutral-600">
              The layout borrows the glass pill precision from the reference and stretches it into a complete AIDA landing page.
            </p>
          </div>

          <div className="grid grid-flow-dense grid-cols-1 gap-4 lg:grid-cols-12">
            {bentoCards.map((card) => (
              <article key={card.title} className={`group min-h-[22rem] overflow-hidden rounded-[2rem] border border-white/70 bg-white/72 p-5 shadow-[0_24px_70px_-38px_rgba(0,0,0,0.5)] backdrop-blur-md ${card.className}`}>
                <div data-scale-image className="mb-6 h-48 overflow-hidden rounded-[1.35rem] bg-neutral-900 md:h-56">
                  <img src={card.image} alt="Abstract landing page texture" className="h-full w-full object-cover opacity-90 grayscale mix-blend-luminosity contrast-125 transition-transform duration-700 ease-out group-hover:scale-105" />
                </div>
                <h3 className="max-w-xl text-3xl font-semibold leading-none tracking-[-0.045em] text-neutral-950">{card.title}</h3>
                <p className="mt-4 max-w-2xl text-sm leading-6 text-neutral-600">{card.body}</p>
              </article>
            ))}
          </div>
        </div>
      </section>

      <section className="relative z-10 overflow-hidden py-12">
        <div className="marquee-track flex w-max gap-4 whitespace-nowrap text-[clamp(2.5rem,7vw,7rem)] font-semibold tracking-[-0.08em] text-neutral-950/90">
          {[...marqueeWords, ...marqueeWords].map((word, index) => (
            <span key={`${word}-${index}`} className="px-6">{word}</span>
          ))}
        </div>
      </section>

      <section id="guide" className="relative z-10 px-5 py-32 md:px-10 md:py-48">
        <div className="mx-auto grid max-w-7xl gap-10 lg:grid-cols-[0.85fr_1.15fr] lg:items-start">
          <div className="lg:sticky lg:top-32">
            <h2 className="text-[clamp(2.8rem,5vw,5.8rem)] font-semibold leading-[0.9] tracking-[-0.075em]">Three slices of polish.</h2>
            <p data-reveal-copy className="mt-8 max-w-md text-2xl leading-snug tracking-[-0.04em] text-neutral-900">
              {revealWords.map((word) => (
                <span data-reveal-word key={word} className="mr-2 inline-block text-neutral-950">{word}</span>
              ))}
            </p>
          </div>

          <div className="flex min-h-[34rem] gap-3 overflow-hidden rounded-[2rem] border border-white/70 bg-white/70 p-3 shadow-[0_30px_90px_-44px_rgba(0,0,0,0.5)] backdrop-blur-md">
            {accordionItems.map((item) => (
              <article key={item.title} className="group relative min-w-0 flex-1 overflow-hidden rounded-[1.45rem] bg-neutral-950 transition-[flex-grow] duration-700 ease-[cubic-bezier(0.16,1,0.3,1)] hover:flex-[2.4]">
                <img data-scale-image src={item.image} alt="Polished visual surface" className="absolute inset-0 h-full w-full object-cover opacity-70 grayscale mix-blend-luminosity contrast-125 transition-transform duration-700 ease-out group-hover:scale-105" />
                <div className="absolute inset-0 bg-gradient-to-t from-black/80 via-black/25 to-transparent" />
                <div className="absolute bottom-0 left-0 right-0 p-6 text-white">
                  <h3 className="text-2xl font-semibold tracking-[-0.04em]">{item.title}</h3>
                  <p className="mt-3 max-w-xs text-sm leading-6 text-white/72 opacity-0 transition-opacity duration-500 ease-out group-hover:opacity-100">{item.body}</p>
                </div>
              </article>
            ))}
          </div>
        </div>
      </section>

      <section id="launch" className="relative z-10 px-5 pb-16 pt-32 md:px-10 md:pt-48">
        <div className="mx-auto max-w-7xl overflow-hidden rounded-[2.5rem] bg-neutral-950 px-6 py-14 text-white shadow-[0_50px_140px_-54px_rgba(0,0,0,0.8)] md:px-12 md:py-20">
          <div className="grid gap-12 md:grid-cols-[1fr_auto] md:items-end">
            <div>
              <h2 className="max-w-4xl text-[clamp(3rem,6vw,6.5rem)] font-semibold leading-[0.9] tracking-[-0.08em]">Preview the page. Ship the image.</h2>
              <p className="mt-8 max-w-2xl text-lg leading-8 text-white/68">
                Run the local Docker scripts, inspect the page, then let CI publish `hello-world` for the simple-service GitOps path.
              </p>
            </div>
            <div className="rounded-[1.5rem] border border-white/10 bg-white/[0.06] p-4 font-mono text-sm leading-7 text-white/80">
              <p>npm ci</p>
              <p>npm run build</p>
              <p>npm run docker:smoke</p>
            </div>
          </div>
        </div>
      </section>

      <footer className="relative z-10 flex flex-col gap-4 px-5 pb-10 text-sm text-neutral-500 md:flex-row md:items-center md:justify-between md:px-10">
        <p>Hello World landing surface.</p>
        <div className="flex gap-5">
          <a href="#docs" className="hover:text-neutral-950">Docs</a>
          <a href="#blog" className="hover:text-neutral-950">Blog</a>
          <a href="#launch" className="hover:text-neutral-950">Launch</a>
        </div>
      </footer>
    </main>
  );
}
