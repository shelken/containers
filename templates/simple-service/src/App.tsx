export default function App() {
  return (
    <main className="min-h-screen bg-slate-950 text-slate-100">
      <section className="mx-auto flex min-h-screen w-full max-w-3xl flex-col justify-center px-6 py-16">
        <p className="mb-4 text-sm font-medium uppercase tracking-[0.35em] text-cyan-300">
          simple-service
        </p>
        <h1 className="text-4xl font-semibold tracking-tight text-white sm:text-6xl">
          Lightweight frontend service
        </h1>
        <p className="mt-6 max-w-2xl text-lg leading-8 text-slate-300">
          Copy this template, rename the app, replace this screen with one focused
          tool or page, then let CI publish a static container image.
        </p>
        <div className="mt-10 rounded-2xl border border-slate-800 bg-slate-900/70 p-5 text-sm text-slate-300 shadow-2xl shadow-cyan-950/20">
          <p className="font-medium text-slate-100">Starter checklist</p>
          <ul className="mt-3 list-disc space-y-2 pl-5">
            <li>Rename <code>APP</code> in <code>docker-bake.hcl</code>.</li>
            <li>Rename <code>package.json</code> and <code>index.html</code> title.</li>
            <li>Replace <code>src/App.tsx</code> with one focused service.</li>
          </ul>
        </div>
      </section>
    </main>
  );
}
