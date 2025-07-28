import { useState, useEffect } from 'react'
import './App.css'

// Dynamically get the base URL from the current browser location
const API_BASE = `http://api.dns.home:8080/api`

function App() {
  const [records, setRecords] = useState([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showAddForm, setShowAddForm] = useState(false)
  const [newRecord, setNewRecord] = useState({ domain: '', ip: '' })

  // Fetch DNS records
  const fetchRecords = async () => {
    try {
      setLoading(true)
      const response = await fetch(`${API_BASE}/records`)
      const data = await response.json()
      
      if (data.success) {
        setRecords(data.data || [])
        setError('')
      } else {
        setError(data.message || 'Failed to fetch records')
      }
    } catch (err) {
      setError('Failed to connect to API server')
      console.error('Fetch error:', err)
    } finally {
      setLoading(false)
    }
  }

  // Add new DNS record
  const addRecord = async (e) => {
    e.preventDefault()
    
    if (!newRecord.domain || !newRecord.ip) {
      setError('Domain and IP are required')
      return
    }

    try {
      const payload = {
        domain: newRecord.domain.trim(),
        ip: newRecord.ip.trim(),
        ttl: -1 // Always set to -1 (no expiration)
      }

      const response = await fetch(`${API_BASE}/records`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload)
      })

      const data = await response.json()
      
      if (data.success) {
        setNewRecord({ domain: '', ip: '' })
        setShowAddForm(false)
        fetchRecords() // Refresh the list
        setError('')
      } else {
        setError(data.message || 'Failed to add record')
      }
    } catch (err) {
      setError('Failed to add record')
      console.error('Add error:', err)
    }
  }

  // Delete DNS record
  const deleteRecord = async (domain) => {
    if (!confirm(`Are you sure you want to delete the record for "${domain}"?`)) {
      return
    }

    try {
      const response = await fetch(`${API_BASE}/records/${encodeURIComponent(domain)}`, {
        method: 'DELETE'
      })

      const data = await response.json()
      
      if (data.success) {
        fetchRecords() // Refresh the list
        setError('')
      } else {
        setError(data.message || 'Failed to delete record')
      }
    } catch (err) {
      setError('Failed to delete record')
      console.error('Delete error:', err)
    }
  }

  useEffect(() => {
    fetchRecords()
  }, [])

  return (
    <div className="app">
      <div className="background-pattern"></div>
      
      <header className="header">
        <div className="header-content">
          <div className="logo">
            <div className="logo-icon">🌐</div>
            <div className="logo-text">
              <h1>DNS Manager</h1>
              <p>Manage your DNS records with ease</p>
            </div>
          </div>
        </div>
      </header>

      <main className="main">
        <div className="container">
          {error && (
            <div className="alert alert-error">
              <div className="alert-icon">⚠️</div>
              <div className="alert-content">
                <span>{error}</span>
                <button onClick={() => setError('')} className="alert-close">×</button>
              </div>
            </div>
          )}

          <div className="toolbar">
            <div className="toolbar-left">
              <h2 className="page-title">DNS Records</h2>
              <span className="record-count">{records.length} records</span>
            </div>
            <div className="toolbar-right">
              <button 
                onClick={fetchRecords}
                className="btn btn-secondary"
                disabled={loading}
              >
                <span className="btn-icon">🔄</span>
                {loading ? 'Refreshing...' : 'Refresh'}
              </button>
              <button 
                onClick={() => setShowAddForm(!showAddForm)}
                className="btn btn-primary"
              >
                <span className="btn-icon">+</span>
                {showAddForm ? 'Cancel' : 'Add Record'}
              </button>
            </div>
          </div>

          {showAddForm && (
            <div className="card add-form-card">
              <div className="card-header">
                <h3>Add New DNS Record</h3>
                <p>Create a new domain to IP address mapping</p>
              </div>
              <form onSubmit={addRecord} className="add-form">
                <div className="form-grid">
                  <div className="form-field">
                    <label htmlFor="domain">Domain Name</label>
                    <input
                      type="text"
                      id="domain"
                      placeholder="example.local"
                      value={newRecord.domain}
                      onChange={(e) => setNewRecord({...newRecord, domain: e.target.value})}
                      required
                    />
                  </div>
                  <div className="form-field">
                    <label htmlFor="ip">IP Address</label>
                    <input
                      type="text"
                      id="ip"
                      placeholder="192.168.1.100"
                      value={newRecord.ip}
                      onChange={(e) => setNewRecord({...newRecord, ip: e.target.value})}
                      required
                    />
                  </div>
                </div>
                <div className="form-actions">
                  <button type="submit" className="btn btn-success">
                    <span className="btn-icon">✓</span>
                    Add Record
                  </button>
                  <button 
                    type="button" 
                    onClick={() => setShowAddForm(false)}
                    className="btn btn-outline"
                  >
                    Cancel
                  </button>
                </div>
              </form>
            </div>
          )}

          <div className="records-section">
            {loading ? (
              <div className="loading-state">
                <div className="loading-spinner"></div>
                <p>Loading DNS records...</p>
              </div>
            ) : records.length === 0 ? (
              <div className="empty-state">
                <div className="empty-icon">📝</div>
                <h3>No DNS records found</h3>
                <p>Get started by adding your first DNS record</p>
                <button 
                  onClick={() => setShowAddForm(true)}
                  className="btn btn-primary"
                >
                  <span className="btn-icon">+</span>
                  Add First Record
                </button>
              </div>
            ) : (
              <div className="records-grid">
                {records.map((record, index) => (
                  <div key={index} className="record-card">
                    <div className="record-header">
                      <div className="record-status active"></div>
                      <div className="record-type">A Record</div>
                    </div>
                    <div className="record-content">
                      <div className="record-field">
                        <label>Domain</label>
                        <div className="record-value domain-value">{record.domain}</div>
                      </div>
                      <div className="record-field">
                        <label>Points to</label>
                        <div className="record-value ip-value">{record.ip}</div>
                      </div>
                    </div>
                    <div className="record-actions">
                      <button
                        onClick={() => deleteRecord(record.domain)}
                        className="btn btn-danger btn-small"
                        title="Delete record"
                      >
                        <span className="btn-icon">🗑️</span>
                        Delete
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        </div>
      </main>
    </div>
  )
}

export default App
