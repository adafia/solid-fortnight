import { BrowserRouter, Routes, Route, NavLink, Link, useParams } from 'react-router-dom';
import { LayoutDashboard, Flag, Settings, Plus, Box, Layers, Terminal, X, Save, Trash2, ArrowRight } from 'lucide-react';
import { useState, useEffect } from 'react';
import { Project, Environment, FeatureFlag, FlagConfig, Rule } from './types/api';

// In production/docker, Caddy handles the proxying from /api to the gateway.
// When running 'bun run dev' locally, you might need a different base or a proxy setting in vite.config.
const API_BASE = window.location.origin.includes('5173') 
  ? 'http://localhost:8080/api/v1/management' 
  : '/api/v1/management';

function ProjectsView() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [newName, setNewName] = useState('');
  const [newDesc, setNewDesc] = useState('');

  const [error, setError] = useState<string | null>(null);

  const fetchProjects = () => {
    setLoading(true);
    setError(null);
    fetch(`${API_BASE}/projects/`)
      .then(res => {
        if (!res.ok) throw new Error("Failed to fetch");
        return res.json();
      })
      .then(data => {
        setProjects(Array.isArray(data) ? data : []);
        setLoading(false);
      })
      .catch(err => {
        console.error("Failed to fetch projects", err);
        setError("Failed to load projects. Is the gateway running?");
        setProjects([]);
        setLoading(false);
      });
  };

  useEffect(() => {
    fetchProjects();
  }, []);

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    try {
      const res = await fetch(`${API_BASE}/projects/`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: newName, description: newDesc })
      });
      if (res.ok) {
        setShowModal(false);
        setNewName('');
        setNewDesc('');
        fetchProjects();
      } else {
        const errText = await res.text();
        setError(`Failed to create project: ${errText}`);
      }
    } catch (err) {
      setError("Network error. Is the gateway running?");
    }
  };

  return (
    <div>
      <div className="header">
        <h1>Projects</h1>
        <button className="btn btn-primary" onClick={() => setShowModal(true)}>
          <Plus size={18} /> New Project
        </button>
      </div>

      {error && (
        <div className="card" style={{ background: '#fef2f2', color: '#991b1b', marginBottom: '2rem', border: '1px solid #fecaca' }}>
          {error}
        </div>
      )}

      {showModal && (
        <div className="modal-overlay">
          <div className="modal">
            <div className="header">
              <h2>Create New Project</h2>
              <button className="btn" onClick={() => setShowModal(false)}><X size={20}/></button>
            </div>
            <form onSubmit={handleCreate}>
              <div className="form-group">
                <label htmlFor="projectName">Project Name</label>
                <input id="projectName" required className="form-control" value={newName} onChange={e => setNewName(e.target.value)} placeholder="e.g. Mobile App" />
              </div>
              <div className="form-group">
                <label htmlFor="projectDesc">Description</label>
                <textarea id="projectDesc" className="form-control" value={newDesc} onChange={e => setNewDesc(e.target.value)} placeholder="What is this project for?" />
              </div>
              <div style={{ display: 'flex', gap: '1rem', justifyContent: 'flex-end', marginTop: '2rem' }}>
                <button type="button" className="btn" onClick={() => setShowModal(false)}>Cancel</button>
                <button type="submit" className="btn btn-primary">Create Project</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {loading ? (
        <p>Loading projects...</p>
      ) : (
        <div className="grid">
          {projects.map(p => (
            <Link key={p.id} to={`/projects/${p.id}`} className="card" style={{ textDecoration: 'none', color: 'inherit' }}>
              <Box size={24} color="#2563eb" style={{ marginBottom: '1rem' }} />
              <h2>{p.name}</h2>
              <p>{p.description || 'No description'}</p>
              <div style={{ marginTop: '1rem', fontSize: '0.75rem', color: '#94a3b8' }}>
                Created {new Date(p.created_at).toLocaleDateString()}
              </div>
            </Link>
          ))}
          {projects.length === 0 && (
            <div className="card" style={{ gridColumn: '1/-1', textAlign: 'center', padding: '3rem' }}>
              <p>No projects found. Create your first project to get started!</p>
            </div>
          )}
          </div>

      )}
    </div>
  );
}

function FlagDetailView() {
  const { projectId, flagKey } = useParams();
  const [project, setProject] = useState<Project | null>(null);
  const [envs, setEnvs] = useState<Environment[]>([]);
  const [activeEnv, setActiveEnv] = useState<string>('');
  const [config, setConfig] = useState<FlagConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    Promise.all([
      fetch(`${API_BASE}/projects/${projectId}`).then(res => res.json()),
      fetch(`${API_BASE}/projects/${projectId}/environments`).then(res => res.json())
    ]).then(([projData, envData]) => {
      setProject(projData);
      setEnvs(Array.isArray(envData) ? envData : []);
      if (envData.length > 0) setActiveEnv(envData[0].id);
      setLoading(false);
    });
  }, [projectId]);

  useEffect(() => {
    if (activeEnv && flagKey) {
      setLoading(true);
      fetch(`${API_BASE}/projects/${projectId}/flags/${flagKey}/environments/${activeEnv}`)
        .then(res => res.json())
        .then(data => {
          setConfig(data || { enabled: false, default_value: false, rules: [] });
          setLoading(false);
        })
        .catch(() => {
          setConfig({ enabled: false, default_value: false, rules: [] });
          setLoading(false);
        });
    }
  }, [activeEnv, flagKey, projectId]);

  const handleSave = async () => {
    setSaving(true);
    try {
      await fetch(`${API_BASE}/projects/${projectId}/flags/${flagKey}/environments/${activeEnv}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config)
      });
    } catch (err) {
      console.error("Failed to save config", err);
    } finally {
      setSaving(false);
    }
  };

  const addRule = () => {
    if (!config) return;
    const newRule: Rule = {
      name: `Rule ${config.rules.length + 1}`,
      priority: config.rules.length + 1,
      clauses: [{ attribute: 'user_id', operator: 'IN', values: [] }],
      value: true
    };
    setConfig({ ...config, rules: [...config.rules, newRule] });
  };

  if (loading && !config) return <p>Loading flag configuration...</p>;
  if (!project) return <p>Project not found.</p>;

  return (
    <div>
      <div className="header">
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '0.5rem' }}>
            <Link to={`/projects/${projectId}`} style={{ color: 'var(--text-muted)', textDecoration: 'none' }}>{project.name}</Link>
            <span style={{ color: 'var(--border)' }}>/</span>
            <h1 style={{ fontSize: '1.5rem' }}>{flagKey}</h1>
          </div>
        </div>
        <button className="btn btn-primary" onClick={handleSave} disabled={saving}>
          <Save size={18} /> {saving ? 'Saving...' : 'Save Configuration'}
        </button>
      </div>

      <div style={{ display: 'flex', gap: '1rem', marginBottom: '2rem', borderBottom: '1px solid var(--border)', paddingBottom: '1rem' }}>
        {envs.map(e => (
          <button 
            key={e.id} 
            className={`btn ${activeEnv === e.id ? 'btn-primary' : ''}`}
            onClick={() => setActiveEnv(e.id)}
            style={{ padding: '0.5rem 1rem' }}
          >
            {e.name}
          </button>
        ))}
      </div>

      {config && (
        <div style={{ maxWidth: '800px' }}>
          <div className="card" style={{ marginBottom: '2rem', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div>
              <h2 style={{ margin: 0 }}>Enable Flag</h2>
              <p>Turn this flag on or off for this environment.</p>
            </div>
            <label className="switch">
              <input type="checkbox" checked={config.enabled} onChange={e => setConfig({ ...config, enabled: e.target.checked })} />
              <span className="slider"></span>
            </label>
          </div>

          <div className="card" style={{ marginBottom: '2rem' }}>
            <h2>Default Value</h2>
            <p style={{ marginBottom: '1rem' }}>The value returned if no targeting rules match.</p>
            <select 
              className="form-control" 
              value={String(config.default_value)} 
              onChange={e => setConfig({ ...config, default_value: e.target.value === 'true' ? true : e.target.value === 'false' ? false : e.target.value })}
            >
              <option value="true">True</option>
              <option value="false">False</option>
            </select>
          </div>

          <div className="header">
            <h2>Targeting Rules</h2>
            <button className="btn" onClick={addRule}><Plus size={18} /> Add Rule</button>
          </div>

          {config.rules.map((rule, rIdx) => (
            <div key={rIdx} className="rule-card">
              <div className="rule-header">
                <input 
                  className="form-control" 
                  style={{ width: 'auto', background: 'transparent', border: 'none', fontWeight: 600 }}
                  value={rule.name}
                  onChange={e => {
                    const newRules = [...config.rules];
                    newRules[rIdx].name = e.target.value;
                    setConfig({ ...config, rules: newRules });
                  }}
                />
                <button className="btn" style={{ color: 'var(--danger)' }} onClick={() => {
                  const newRules = config.rules.filter((_, i) => i !== rIdx);
                  setConfig({ ...config, rules: newRules });
                }}>
                  <Trash2 size={16} />
                </button>
              </div>
              <div className="rule-body">
                {rule.clauses.map((clause, cIdx) => (
                  <div key={cIdx} className="clause-row">
                    <input 
                      className="form-control" 
                      style={{ width: '150px' }} 
                      value={clause.attribute}
                      onChange={e => {
                        const newRules = [...config.rules];
                        newRules[rIdx].clauses[cIdx].attribute = e.target.value;
                        setConfig({ ...config, rules: newRules });
                      }}
                    />
                    <select 
                      className="form-control" 
                      style={{ width: '150px' }}
                      value={clause.operator}
                      onChange={e => {
                        const newRules = [...config.rules];
                        newRules[rIdx].clauses[cIdx].operator = e.target.value as any;
                        setConfig({ ...config, rules: newRules });
                      }}
                    >
                      <option value="EQUALS">Equals</option>
                      <option value="IN">In List</option>
                      <option value="CONTAINS">Contains</option>
                    </select>
                    <input 
                      className="form-control" 
                      placeholder="Values (comma separated)"
                      value={clause.values.join(',')}
                      onChange={e => {
                        const newRules = [...config.rules];
                        newRules[rIdx].clauses[cIdx].values = e.target.value.split(',').map(v => v.trim());
                        setConfig({ ...config, rules: newRules });
                      }}
                    />
                  </div>
                ))}
                <div className="rule-result">
                  <ArrowRight size={20} color="var(--text-muted)" />
                  <span>Return</span>
                  <select 
                    className="form-control" 
                    style={{ width: '120px' }}
                    value={String(rule.value)}
                    onChange={e => {
                      const newRules = [...config.rules];
                      newRules[rIdx].value = e.target.value === 'true' ? true : e.target.value === 'false' ? false : e.target.value;
                      setConfig({ ...config, rules: newRules });
                    }}
                  >
                    <option value="true">True</option>
                    <option value="false">False</option>
                  </select>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function ProjectDetailView() {
  const { id } = useParams();
  const [project, setProject] = useState<Project | null>(null);
  const [envs, setEnvs] = useState<Environment[]>([]);
  const [flags, setFlags] = useState<FeatureFlag[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    Promise.all([
      fetch(`${API_BASE}/projects/${id}`).then(res => res.json()),
      fetch(`${API_BASE}/projects/${id}/environments`).then(res => res.json()),
      fetch(`${API_BASE}/flags/?project_id=${id}`).then(res => res.json())
    ]).then(([projData, envData, flagData]) => {
      setProject(projData);
      setEnvs(Array.isArray(envData) ? envData : []);
      setFlags(Array.isArray(flagData) ? flagData : []);
      setLoading(false);
    }).catch(err => {
      console.error("Failed to fetch project details", err);
      setLoading(false);
    });
  }, [id]);

  if (loading) return <p>Loading details...</p>;
  if (!project) return <p>Project not found.</p>;

  return (
    <div>
      <div className="header">
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '1rem', marginBottom: '0.5rem' }}>
            <Link to="/projects" style={{ color: 'var(--text-muted)', textDecoration: 'none' }}>Projects</Link>
            <span style={{ color: 'var(--border)' }}>/</span>
            <h1 style={{ fontSize: '1.5rem' }}>{project.name}</h1>
          </div>
          <p>{project.description}</p>
        </div>
        <div style={{ display: 'flex', gap: '1rem' }}>
          <button className="btn"><Settings size={18} /> Settings</button>
          <button className="btn btn-primary"><Plus size={18} /> New Flag</button>
        </div>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '1fr 300px', gap: '2rem', marginTop: '3rem' }}>
        <div>
          <h2 style={{ marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <Flag size={20} /> Feature Flags
          </h2>
          <div className="card" style={{ padding: 0 }}>
            {flags.length === 0 ? (
              <div style={{ padding: '3rem', textAlign: 'center' }}>
                <p>No flags created yet for this project.</p>
              </div>
            ) : (
              <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                <thead>
                  <tr style={{ textAlign: 'left', borderBottom: '1px solid var(--border)' }}>
                    <th style={{ padding: '1rem' }}>Flag Name</th>
                    <th style={{ padding: '1rem' }}>Key</th>
                    <th style={{ padding: '1rem' }}>Type</th>
                  </tr>
                </thead>
                <tbody>
                  {flags.map(f => (
                    <tr key={f.id} style={{ borderBottom: '1px solid var(--border)' }}>
                      <td style={{ padding: '1rem', fontWeight: 500 }}>
                        <Link to={`/projects/${id}/flags/${f.key}`} style={{ textDecoration: 'none', color: 'inherit' }}>{f.name}</Link>
                      </td>
                      <td style={{ padding: '1rem' }}><code>{f.key}</code></td>
                      <td style={{ padding: '1rem' }}>
                        <span className="badge badge-gray">{f.type}</span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>

        <div>
          <h2 style={{ marginBottom: '1.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <Layers size={20} /> Environments
          </h2>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
            {envs.map(e => (
              <div key={e.id} className="card" style={{ padding: '1rem' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ fontWeight: 600 }}>{e.name}</span>
                  <span className="badge badge-blue">{e.key}</span>
                </div>
              </div>
            ))}
            <button className="btn" style={{ width: '100%', justifyContent: 'center', borderStyle: 'dashed', borderColor: 'var(--border)' }}>
              <Plus size={16} /> Add Environment
            </button>
          </div>

          <h2 style={{ margin: '2rem 0 1rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <Terminal size={20} /> SDK Keys
          </h2>
          <div className="card" style={{ background: '#1e293b', color: '#f8fafc', fontSize: '0.875rem' }}>
            <code>PROJECT_ID={project.id}</code>
          </div>
        </div>
      </div>
    </div>
  );
}

function App() {
  return (
    <BrowserRouter>
      <div className="layout">
        <aside className="sidebar">
          <div style={{ padding: '0 1rem 2rem', fontWeight: 800, fontSize: '1.25rem', color: '#2563eb' }}>
            Solid Fortnight
          </div>
          <nav>
            <NavLink to="/" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
              <LayoutDashboard size={20} /> Dashboard
            </NavLink>
            <NavLink to="/projects" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
              <Box size={20} /> Projects
            </NavLink>
            <NavLink to="/flags" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
              <Flag size={20} /> Global Flags
            </NavLink>
            <NavLink to="/settings" className={({ isActive }) => `nav-item ${isActive ? 'active' : ''}`}>
              <Settings size={20} /> Settings
            </NavLink>
          </nav>
        </aside>

        <main className="main">
          <Routes>
            <Route path="/" element={<div><h1>Welcome to Solid Fortnight</h1><p>Select a project to manage your feature flags.</p></div>} />
            <Route path="/projects" element={<ProjectsView />} />
            <Route path="/projects/:id" element={<ProjectDetailView />} />
            <Route path="/projects/:projectId/flags/:flagKey" element={<FlagDetailView />} />
            <Route path="/flags" element={<div><h1>Global Flags</h1></div>} />
            <Route path="/settings" element={<div><h1>Settings</h1></div>} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  );
}

export default App;
