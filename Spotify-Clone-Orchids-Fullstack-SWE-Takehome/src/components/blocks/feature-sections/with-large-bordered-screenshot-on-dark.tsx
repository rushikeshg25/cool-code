export default function WithLargeBorderedScreenshotOnDark() {
    return (
      <div className="overflow-hidden bg-gray-900 py-24 sm:py-32">
        <div className="mx-auto max-w-7xl px-6 lg:px-8">
          <p className="max-w-2xl text-5xl font-semibold tracking-tight text-pretty text-white sm:text-6xl sm:text-balance">
            Everything you need to deploy your app
          </p>
          <div className="relative mt-16 aspect-2432/1442 h-144 sm:h-auto sm:w-[calc(var(--container-7xl)-calc(var(--spacing)*16))]">
            <div className="absolute -inset-2 rounded-[calc(var(--radius-xl)+calc(var(--spacing)*2))] bg-white/2.5 shadow-xs ring-1 ring-white/10" />
  
            <img
              alt=""
              src="https://tailwindcss.com/plus-assets/img/component-images/dark-project-app-screenshot.png"
              className="h-full rounded-xl shadow-2xl ring-1 ring-white/10"
            />
          </div>
        </div>
      </div>
    )
  }
  