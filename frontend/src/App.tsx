import { useEffect, useState } from "react";

function App() {
  const [argoApps, setArgoApps] = useState<Array<string> | null>(null);
  const [reload, setReload] = useState<boolean>(false);
  const [applicationName, setApplicationName] = useState<string>("");
  const [repoURL, setRepoURL] = useState<string>("");
  const [clusterURL, setClusterURL] = useState<string>("");
  const [path, setPath] = useState<string>("");

  useEffect(() => {
    getDeployments();
  }, [reload]);

  async function getDeployments() {
    const response = await fetch("/argo/list");
    const data = await response.json();

    setArgoApps(data);
  }

  function triggerReload() {
    setReload((prev) => !prev);
  }

  async function deleteArgoApplication(argoAppName: string) {
    await fetch(`/argo/delete/${argoAppName}`, {
      method: "DELETE",
    });
    triggerReload();
  }

  async function createArgoApplication(values: {
    applicationName: string;
    repositoryURL: string;
    clusterURL: string;
    path: string;
  }) {
    triggerReload();
    await fetch(`/argo/create`, {
      method: "POST",
      body: JSON.stringify(values),
      headers: {
        "Content-Type": "application/json",
      },
    });
  }

  return (
    <>
      <div className="mt-36 flex flex-col items-center justify-center">
        <p className="mb-24 font-bold text-3xl">Argo Application Manager</p>
        <form
          className="flex space-x-4 mb-12"
          onSubmit={async () => {
            await createArgoApplication({
              applicationName,
              repositoryURL: repoURL,
              clusterURL,
              path,
            });
          }}
        >
          <input
            placeholder="Application Name"
            type="text"
            className="border border-2 px-2"
            onChange={(event) => {
              setApplicationName(event.target.value);
            }}
          />
          <input
            placeholder="Repo URL"
            type="text"
            className="border border-2 px-2"
            onChange={(event) => {
              setRepoURL(event.target.value);
            }}
          />
          <input
            placeholder="Cluster URL"
            type="text"
            className="border border-2 px-2"
            onChange={(event) => {
              setClusterURL(event.target.value);
            }}
          />
          <input
            placeholder="Path"
            type="text"
            className="border border-2 px-2"
            onChange={(event) => {
              setPath(event.target.value);
            }}
          />
          <button type="submit" className="bg-green-300 px-4 py-2 rounded-md">
            Create
          </button>
        </form>
        {argoApps?.map((argoApp, i) => (
          <>
            <div
              key={i}
              className="flex items-center space-x-4 border border-1 border-black p-4"
            >
              <div>{i + 1}.</div>
              <div>{argoApp}</div>
              <button
                className="bg-red-300 px-4 py-2 rounded-md"
                onClick={async (e) => {
                  e.preventDefault();
                  await deleteArgoApplication(argoApp);
                  triggerReload();
                }}
              >
                Delete
              </button>
            </div>
          </>
        ))}
      </div>
    </>
  );
}

export default App;
