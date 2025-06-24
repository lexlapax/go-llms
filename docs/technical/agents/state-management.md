# Agent State Management: State and Data Flow

> **[Project Root](/) / [Documentation](../..) / [Technical Documentation](../../technical) / [Agents](../../technical/agents) / State Management**

Comprehensive guide to agent state management in Go-LLMs, covering state architecture, persistence strategies, data flow patterns, state synchronization, versioning, and advanced state management techniques for building robust, stateful agents.

## State Management Architecture

### Core State Interfaces

```go
// AgentState represents the state of an agent
type AgentState interface {
    // Basic state operations
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}) error
    Delete(key string) error
    Has(key string) bool
    
    // Bulk operations
    GetAll() map[string]interface{}
    SetAll(data map[string]interface{}) error
    Clear() error
    
    // State metadata
    GetMetadata() StateMetadata
    SetMetadata(metadata StateMetadata) error
    
    // Versioning
    GetVersion() int64
    CreateSnapshot() (StateSnapshot, error)
    RestoreSnapshot(snapshot StateSnapshot) error
    
    // Serialization
    Marshal() ([]byte, error)
    Unmarshal(data []byte) error
    
    // Validation
    Validate() error
    GetSchema() StateSchema
}

// StateManager handles state operations and persistence
type StateManager interface {
    // State lifecycle
    CreateState(agentID string, initialData map[string]interface{}) (AgentState, error)
    LoadState(agentID string) (AgentState, error)
    SaveState(agentID string, state AgentState) error
    DeleteState(agentID string) error
    
    // State queries
    ListStates() ([]string, error)
    StateExists(agentID string) bool
    
    // Transactions
    BeginTransaction(agentID string) (StateTransaction, error)
    CommitTransaction(tx StateTransaction) error
    RollbackTransaction(tx StateTransaction) error
    
    // Backup and restore
    BackupState(agentID string) (StateBackup, error)
    RestoreState(agentID string, backup StateBackup) error
    
    // Monitoring
    GetStateMetrics(agentID string) StateMetrics
    WatchState(agentID string) (<-chan StateEvent, error)
}

// StatePersistence handles different storage backends
type StatePersistence interface {
    // Basic persistence
    Store(key string, data []byte) error
    Load(key string) ([]byte, error)
    Delete(key string) error
    Exists(key string) bool
    
    // Batch operations
    StoreBatch(items map[string][]byte) error
    LoadBatch(keys []string) (map[string][]byte, error)
    
    // Atomic operations
    CompareAndSwap(key string, expected, new []byte) (bool, error)
    AtomicIncrement(key string, delta int64) (int64, error)
    
    // Cleanup
    Cleanup() error
    GetStats() PersistenceStats
}

type StateMetadata struct {
    AgentID     string                 `json:"agent_id"`
    Version     int64                  `json:"version"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
    Size        int64                  `json:"size"`
    Checksum    string                 `json:"checksum"`
    Tags        map[string]string      `json:"tags,omitempty"`
    Attributes  map[string]interface{} `json:"attributes,omitempty"`
}

type StateSnapshot struct {
    ID        string                 `json:"id"`
    AgentID   string                 `json:"agent_id"`
    Version   int64                  `json:"version"`
    Data      map[string]interface{} `json:"data"`
    Metadata  StateMetadata          `json:"metadata"`
    CreatedAt time.Time              `json:"created_at"`
}

type StateSchema struct {
    Fields      map[string]FieldSchema `json:"fields"`
    Required    []string               `json:"required"`
    Constraints []Constraint           `json:"constraints"`
    Version     string                 `json:"version"`
}

type FieldSchema struct {
    Type        string      `json:"type"`
    Description string      `json:"description"`
    Default     interface{} `json:"default,omitempty"`
    Validation  *Validation `json:"validation,omitempty"`
}
```

### Default State Implementation

```go
// MemoryAgentState implements AgentState using in-memory storage
type MemoryAgentState struct {
    data      map[string]interface{}
    metadata  StateMetadata
    version   int64
    schema    StateSchema
    snapshots []StateSnapshot
    mu        sync.RWMutex
    
    // Event handling
    eventBus  EventBus
    listeners []StateListener
    
    // Validation
    validator StateValidator
    
    // Metrics
    metrics *StateMetrics
    logger  *zap.Logger
}

func NewMemoryAgentState(agentID string, schema StateSchema) *MemoryAgentState {
    return &MemoryAgentState{
        data:      make(map[string]interface{}),
        metadata: StateMetadata{
            AgentID:   agentID,
            Version:   1,
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
            Tags:      make(map[string]string),
            Attributes: make(map[string]interface{}),
        },
        version:   1,
        schema:    schema,
        snapshots: make([]StateSnapshot, 0),
        eventBus:  NewEventBus(),
        listeners: make([]StateListener, 0),
        validator: NewStateValidator(schema),
        metrics:   NewStateMetrics(agentID),
        logger:    zap.NewNop(),
    }
}

func (s *MemoryAgentState) Get(key string) (interface{}, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    value, exists := s.data[key]
    
    // Record access metrics
    s.metrics.RecordAccess(key, "get")
    
    return value, exists
}

func (s *MemoryAgentState) Set(key string, value interface{}) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Validate field
    if err := s.validateField(key, value); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    // Store old value for event
    oldValue, existed := s.data[key]
    
    // Set new value
    s.data[key] = value
    s.version++
    s.metadata.UpdatedAt = time.Now()
    s.metadata.Version = s.version
    
    // Update size and checksum
    s.updateMetadata()
    
    // Emit event
    event := StateEvent{
        Type:     StateEventSet,
        AgentID:  s.metadata.AgentID,
        Key:      key,
        Value:    value,
        OldValue: oldValue,
        Version:  s.version,
        Timestamp: time.Now(),
    }
    
    s.emitEvent(event)
    
    // Record metrics
    s.metrics.RecordAccess(key, "set")
    s.metrics.RecordStateChange()
    
    s.logger.Debug("State updated",
        zap.String("agent_id", s.metadata.AgentID),
        zap.String("key", key),
        zap.Int64("version", s.version),
        zap.Bool("existed", existed),
    )
    
    return nil
}

func (s *MemoryAgentState) Delete(key string) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    oldValue, existed := s.data[key]
    if !existed {
        return fmt.Errorf("key %s does not exist", key)
    }
    
    delete(s.data, key)
    s.version++
    s.metadata.UpdatedAt = time.Now()
    s.metadata.Version = s.version
    
    // Update metadata
    s.updateMetadata()
    
    // Emit event
    event := StateEvent{
        Type:     StateEventDelete,
        AgentID:  s.metadata.AgentID,
        Key:      key,
        OldValue: oldValue,
        Version:  s.version,
        Timestamp: time.Now(),
    }
    
    s.emitEvent(event)
    
    // Record metrics
    s.metrics.RecordAccess(key, "delete")
    s.metrics.RecordStateChange()
    
    return nil
}

func (s *MemoryAgentState) CreateSnapshot() (StateSnapshot, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    // Deep copy data
    dataCopy := make(map[string]interface{})
    for k, v := range s.data {
        dataCopy[k] = deepCopy(v)
    }
    
    snapshot := StateSnapshot{
        ID:        generateSnapshotID(),
        AgentID:   s.metadata.AgentID,
        Version:   s.version,
        Data:      dataCopy,
        Metadata:  s.metadata,
        CreatedAt: time.Now(),
    }
    
    // Store snapshot
    s.snapshots = append(s.snapshots, snapshot)
    
    // Limit snapshot history
    maxSnapshots := 10
    if len(s.snapshots) > maxSnapshots {
        s.snapshots = s.snapshots[len(s.snapshots)-maxSnapshots:]
    }
    
    s.logger.Info("State snapshot created",
        zap.String("agent_id", s.metadata.AgentID),
        zap.String("snapshot_id", snapshot.ID),
        zap.Int64("version", s.version),
    )
    
    return snapshot, nil
}

func (s *MemoryAgentState) RestoreSnapshot(snapshot StateSnapshot) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Validate snapshot
    if snapshot.AgentID != s.metadata.AgentID {
        return fmt.Errorf("snapshot agent ID mismatch")
    }
    
    // Backup current state
    currentSnapshot, err := s.createSnapshotLocked()
    if err != nil {
        return fmt.Errorf("failed to backup current state: %w", err)
    }
    
    // Restore data
    s.data = make(map[string]interface{})
    for k, v := range snapshot.Data {
        s.data[k] = deepCopy(v)
    }
    
    s.version++
    s.metadata.UpdatedAt = time.Now()
    s.metadata.Version = s.version
    
    // Update metadata
    s.updateMetadata()
    
    // Emit event
    event := StateEvent{
        Type:      StateEventRestore,
        AgentID:   s.metadata.AgentID,
        Version:   s.version,
        Timestamp: time.Now(),
        Metadata: map[string]interface{}{
            "snapshot_id":      snapshot.ID,
            "snapshot_version": snapshot.Version,
            "backup_snapshot":  currentSnapshot.ID,
        },
    }
    
    s.emitEvent(event)
    
    s.logger.Info("State restored from snapshot",
        zap.String("agent_id", s.metadata.AgentID),
        zap.String("snapshot_id", snapshot.ID),
        zap.Int64("snapshot_version", snapshot.Version),
        zap.Int64("new_version", s.version),
    )
    
    return nil
}

func (s *MemoryAgentState) validateField(key string, value interface{}) error {
    if s.validator != nil {
        return s.validator.ValidateField(key, value, s.schema)
    }
    return nil
}

func (s *MemoryAgentState) updateMetadata() {
    // Calculate size
    size := int64(0)
    for k, v := range s.data {
        size += int64(len(k))
        if str, ok := v.(string); ok {
            size += int64(len(str))
        } else {
            size += 64 // Estimate for other types
        }
    }
    s.metadata.Size = size
    
    // Calculate checksum
    hash := sha256.New()
    keys := make([]string, 0, len(s.data))
    for k := range s.data {
        keys = append(keys, k)
    }
    sort.Strings(keys)
    
    for _, k := range keys {
        hash.Write([]byte(k))
        hash.Write([]byte(fmt.Sprintf("%v", s.data[k])))
    }
    
    s.metadata.Checksum = fmt.Sprintf("%x", hash.Sum(nil))
}

func (s *MemoryAgentState) emitEvent(event StateEvent) {
    // Notify listeners
    for _, listener := range s.listeners {
        go listener.OnStateChange(event)
    }
    
    // Publish to event bus
    s.eventBus.Publish(event)
}

func (s *MemoryAgentState) createSnapshotLocked() (StateSnapshot, error) {
    // This is called when already holding the lock
    dataCopy := make(map[string]interface{})
    for k, v := range s.data {
        dataCopy[k] = deepCopy(v)
    }
    
    return StateSnapshot{
        ID:        generateSnapshotID(),
        AgentID:   s.metadata.AgentID,
        Version:   s.version,
        Data:      dataCopy,
        Metadata:  s.metadata,
        CreatedAt: time.Now(),
    }, nil
}

// State event handling
type StateEvent struct {
    Type      StateEventType         `json:"type"`
    AgentID   string                 `json:"agent_id"`
    Key       string                 `json:"key,omitempty"`
    Value     interface{}            `json:"value,omitempty"`
    OldValue  interface{}            `json:"old_value,omitempty"`
    Version   int64                  `json:"version"`
    Timestamp time.Time              `json:"timestamp"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type StateEventType string

const (
    StateEventSet     StateEventType = "set"
    StateEventDelete  StateEventType = "delete"
    StateEventClear   StateEventType = "clear"
    StateEventRestore StateEventType = "restore"
    StateEventSync    StateEventType = "sync"
)

type StateListener interface {
    OnStateChange(event StateEvent)
}

// Utility functions
func deepCopy(src interface{}) interface{} {
    // Implementation depends on the complexity of data structures
    // For simple types, we can use JSON marshaling/unmarshaling
    data, _ := json.Marshal(src)
    var dst interface{}
    json.Unmarshal(data, &dst)
    return dst
}

func generateSnapshotID() string {
    return fmt.Sprintf("snapshot_%d_%d", time.Now().Unix(), rand.Int63())
}
```

---

## State Persistence Strategies

### File-Based Persistence

```go
// FileStatePersistence implements StatePersistence using file system
type FileStatePersistence struct {
    basePath   string
    fileMode   os.FileMode
    dirMode    os.FileMode
    syncWrites bool
    compression bool
    encryption *EncryptionConfig
    
    // Atomic operations
    lockManager LockManager
    
    // Metrics
    metrics *PersistenceMetrics
    logger  *zap.Logger
}

type EncryptionConfig struct {
    Enabled   bool   `json:"enabled"`
    Algorithm string `json:"algorithm"`
    KeyFile   string `json:"key_file"`
    Key       []byte `json:"-"`
}

func NewFileStatePersistence(basePath string, opts ...FileOption) *FileStatePersistence {
    persistence := &FileStatePersistence{
        basePath:    basePath,
        fileMode:    0644,
        dirMode:     0755,
        syncWrites:  true,
        compression: false,
        lockManager: NewFileLockManager(),
        metrics:     NewPersistenceMetrics("file"),
        logger:      zap.NewNop(),
    }
    
    // Apply options
    for _, opt := range opts {
        opt(persistence)
    }
    
    // Ensure base directory exists
    os.MkdirAll(basePath, persistence.dirMode)
    
    return persistence
}

type FileOption func(*FileStatePersistence)

func WithCompression(enabled bool) FileOption {
    return func(p *FileStatePersistence) {
        p.compression = enabled
    }
}

func WithEncryption(config EncryptionConfig) FileOption {
    return func(p *FileStatePersistence) {
        p.encryption = &config
    }
}

func (p *FileStatePersistence) Store(key string, data []byte) error {
    filePath := p.getFilePath(key)
    
    // Ensure directory exists
    if err := os.MkdirAll(filepath.Dir(filePath), p.dirMode); err != nil {
        return fmt.Errorf("failed to create directory: %w", err)
    }
    
    // Acquire lock for atomic writes
    lock, err := p.lockManager.Lock(key)
    if err != nil {
        return fmt.Errorf("failed to acquire lock: %w", err)
    }
    defer lock.Unlock()
    
    // Process data
    processedData := data
    
    // Apply compression
    if p.compression {
        compressed, err := p.compress(data)
        if err != nil {
            return fmt.Errorf("compression failed: %w", err)
        }
        processedData = compressed
    }
    
    // Apply encryption
    if p.encryption != nil && p.encryption.Enabled {
        encrypted, err := p.encrypt(processedData)
        if err != nil {
            return fmt.Errorf("encryption failed: %w", err)
        }
        processedData = encrypted
    }
    
    // Write to temporary file first
    tempPath := filePath + ".tmp"
    if err := p.writeFile(tempPath, processedData); err != nil {
        return fmt.Errorf("failed to write temp file: %w", err)
    }
    
    // Atomic rename
    if err := os.Rename(tempPath, filePath); err != nil {
        os.Remove(tempPath)
        return fmt.Errorf("failed to rename temp file: %w", err)
    }
    
    // Record metrics
    p.metrics.RecordWrite(key, len(data))
    
    return nil
}

func (p *FileStatePersistence) Load(key string) ([]byte, error) {
    filePath := p.getFilePath(key)
    
    // Check if file exists
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        return nil, fmt.Errorf("key %s not found", key)
    }
    
    // Read file
    data, err := ioutil.ReadFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read file: %w", err)
    }
    
    // Process data
    processedData := data
    
    // Apply decryption
    if p.encryption != nil && p.encryption.Enabled {
        decrypted, err := p.decrypt(processedData)
        if err != nil {
            return nil, fmt.Errorf("decryption failed: %w", err)
        }
        processedData = decrypted
    }
    
    // Apply decompression
    if p.compression {
        decompressed, err := p.decompress(processedData)
        if err != nil {
            return nil, fmt.Errorf("decompression failed: %w", err)
        }
        processedData = decompressed
    }
    
    // Record metrics
    p.metrics.RecordRead(key, len(processedData))
    
    return processedData, nil
}

func (p *FileStatePersistence) writeFile(path string, data []byte) error {
    file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, p.fileMode)
    if err != nil {
        return err
    }
    defer file.Close()
    
    if _, err := file.Write(data); err != nil {
        return err
    }
    
    if p.syncWrites {
        return file.Sync()
    }
    
    return nil
}

func (p *FileStatePersistence) getFilePath(key string) string {
    // Sanitize key for filename
    sanitized := strings.ReplaceAll(key, "/", "_")
    sanitized = strings.ReplaceAll(sanitized, "\\", "_")
    sanitized = strings.ReplaceAll(sanitized, ":", "_")
    
    return filepath.Join(p.basePath, sanitized+".state")
}

func (p *FileStatePersistence) compress(data []byte) ([]byte, error) {
    var buf bytes.Buffer
    writer := gzip.NewWriter(&buf)
    
    if _, err := writer.Write(data); err != nil {
        return nil, err
    }
    
    if err := writer.Close(); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}

func (p *FileStatePersistence) decompress(data []byte) ([]byte, error) {
    reader, err := gzip.NewReader(bytes.NewReader(data))
    if err != nil {
        return nil, err
    }
    defer reader.Close()
    
    return ioutil.ReadAll(reader)
}

func (p *FileStatePersistence) encrypt(data []byte) ([]byte, error) {
    if p.encryption == nil || !p.encryption.Enabled {
        return data, nil
    }
    
    // Simple AES encryption example
    block, err := aes.NewCipher(p.encryption.Key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, data, nil)
    return ciphertext, nil
}

func (p *FileStatePersistence) decrypt(data []byte) ([]byte, error) {
    if p.encryption == nil || !p.encryption.Enabled {
        return data, nil
    }
    
    block, err := aes.NewCipher(p.encryption.Key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return nil, fmt.Errorf("ciphertext too short")
    }
    
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, err
    }
    
    return plaintext, nil
}

// Lock manager for atomic operations
type LockManager interface {
    Lock(key string) (Lock, error)
    TryLock(key string) (Lock, bool, error)
}

type Lock interface {
    Unlock() error
}

type FileLockManager struct {
    locks map[string]*sync.Mutex
    mu    sync.Mutex
}

func NewFileLockManager() *FileLockManager {
    return &FileLockManager{
        locks: make(map[string]*sync.Mutex),
    }
}

func (m *FileLockManager) Lock(key string) (Lock, error) {
    m.mu.Lock()
    
    if _, exists := m.locks[key]; !exists {
        m.locks[key] = &sync.Mutex{}
    }
    
    keyMutex := m.locks[key]
    m.mu.Unlock()
    
    keyMutex.Lock()
    
    return &FileLock{
        manager: m,
        key:     key,
        mutex:   keyMutex,
    }, nil
}

func (m *FileLockManager) TryLock(key string) (Lock, bool, error) {
    m.mu.Lock()
    
    if _, exists := m.locks[key]; !exists {
        m.locks[key] = &sync.Mutex{}
    }
    
    keyMutex := m.locks[key]
    m.mu.Unlock()
    
    if keyMutex.TryLock() {
        return &FileLock{
            manager: m,
            key:     key,
            mutex:   keyMutex,
        }, true, nil
    }
    
    return nil, false, nil
}

type FileLock struct {
    manager *FileLockManager
    key     string
    mutex   *sync.Mutex
}

func (l *FileLock) Unlock() error {
    l.mutex.Unlock()
    return nil
}
```

### Database Persistence

```go
// DatabaseStatePersistence implements StatePersistence using SQL database
type DatabaseStatePersistence struct {
    db          *sql.DB
    tableName   string
    keyColumn   string
    dataColumn  string
    versionColumn string
    timestampColumn string
    
    // Prepared statements
    insertStmt *sql.Stmt
    updateStmt *sql.Stmt
    selectStmt *sql.Stmt
    deleteStmt *sql.Stmt
    existsStmt *sql.Stmt
    
    // Configuration
    compression bool
    encryption  *EncryptionConfig
    
    // Metrics
    metrics *PersistenceMetrics
    logger  *zap.Logger
}

func NewDatabaseStatePersistence(db *sql.DB, tableName string) (*DatabaseStatePersistence, error) {
    persistence := &DatabaseStatePersistence{
        db:              db,
        tableName:       tableName,
        keyColumn:       "key",
        dataColumn:      "data",
        versionColumn:   "version",
        timestampColumn: "updated_at",
        compression:     false,
        metrics:         NewPersistenceMetrics("database"),
        logger:          zap.NewNop(),
    }
    
    // Initialize table
    if err := persistence.initializeTable(); err != nil {
        return nil, fmt.Errorf("failed to initialize table: %w", err)
    }
    
    // Prepare statements
    if err := persistence.prepareStatements(); err != nil {
        return nil, fmt.Errorf("failed to prepare statements: %w", err)
    }
    
    return persistence, nil
}

func (p *DatabaseStatePersistence) initializeTable() error {
    query := fmt.Sprintf(`
        CREATE TABLE IF NOT EXISTS %s (
            %s VARCHAR(255) PRIMARY KEY,
            %s LONGBLOB NOT NULL,
            %s BIGINT NOT NULL DEFAULT 1,
            %s TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
            INDEX idx_timestamp (%s)
        )
    `, p.tableName, p.keyColumn, p.dataColumn, p.versionColumn, p.timestampColumn, p.timestampColumn)
    
    _, err := p.db.Exec(query)
    return err
}

func (p *DatabaseStatePersistence) prepareStatements() error {
    var err error
    
    // Insert statement with ON DUPLICATE KEY UPDATE
    insertQuery := fmt.Sprintf(`
        INSERT INTO %s (%s, %s, %s) VALUES (?, ?, 1)
        ON DUPLICATE KEY UPDATE %s = VALUES(%s), %s = %s + 1
    `, p.tableName, p.keyColumn, p.dataColumn, p.versionColumn,
        p.dataColumn, p.dataColumn, p.versionColumn, p.versionColumn)
    
    p.insertStmt, err = p.db.Prepare(insertQuery)
    if err != nil {
        return fmt.Errorf("failed to prepare insert statement: %w", err)
    }
    
    // Select statement
    selectQuery := fmt.Sprintf(`
        SELECT %s, %s FROM %s WHERE %s = ?
    `, p.dataColumn, p.versionColumn, p.tableName, p.keyColumn)
    
    p.selectStmt, err = p.db.Prepare(selectQuery)
    if err != nil {
        return fmt.Errorf("failed to prepare select statement: %w", err)
    }
    
    // Delete statement
    deleteQuery := fmt.Sprintf(`
        DELETE FROM %s WHERE %s = ?
    `, p.tableName, p.keyColumn)
    
    p.deleteStmt, err = p.db.Prepare(deleteQuery)
    if err != nil {
        return fmt.Errorf("failed to prepare delete statement: %w", err)
    }
    
    // Exists statement
    existsQuery := fmt.Sprintf(`
        SELECT 1 FROM %s WHERE %s = ? LIMIT 1
    `, p.tableName, p.keyColumn)
    
    p.existsStmt, err = p.db.Prepare(existsQuery)
    if err != nil {
        return fmt.Errorf("failed to prepare exists statement: %w", err)
    }
    
    return nil
}

func (p *DatabaseStatePersistence) Store(key string, data []byte) error {
    // Process data
    processedData := data
    
    if p.compression {
        compressed, err := p.compress(data)
        if err != nil {
            return fmt.Errorf("compression failed: %w", err)
        }
        processedData = compressed
    }
    
    if p.encryption != nil && p.encryption.Enabled {
        encrypted, err := p.encrypt(processedData)
        if err != nil {
            return fmt.Errorf("encryption failed: %w", err)
        }
        processedData = encrypted
    }
    
    // Execute insert/update
    result, err := p.insertStmt.Exec(key, processedData)
    if err != nil {
        return fmt.Errorf("failed to store data: %w", err)
    }
    
    // Check if operation was successful
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get affected rows: %w", err)
    }
    
    if rowsAffected == 0 {
        return fmt.Errorf("no rows affected")
    }
    
    // Record metrics
    p.metrics.RecordWrite(key, len(data))
    
    return nil
}

func (p *DatabaseStatePersistence) Load(key string) ([]byte, error) {
    var data []byte
    var version int64
    
    err := p.selectStmt.QueryRow(key).Scan(&data, &version)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("key %s not found", key)
        }
        return nil, fmt.Errorf("failed to load data: %w", err)
    }
    
    // Process data
    processedData := data
    
    if p.encryption != nil && p.encryption.Enabled {
        decrypted, err := p.decrypt(processedData)
        if err != nil {
            return nil, fmt.Errorf("decryption failed: %w", err)
        }
        processedData = decrypted
    }
    
    if p.compression {
        decompressed, err := p.decompress(processedData)
        if err != nil {
            return nil, fmt.Errorf("decompression failed: %w", err)
        }
        processedData = decompressed
    }
    
    // Record metrics
    p.metrics.RecordRead(key, len(processedData))
    
    return processedData, nil
}

func (p *DatabaseStatePersistence) Delete(key string) error {
    result, err := p.deleteStmt.Exec(key)
    if err != nil {
        return fmt.Errorf("failed to delete data: %w", err)
    }
    
    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("failed to get affected rows: %w", err)
    }
    
    if rowsAffected == 0 {
        return fmt.Errorf("key %s not found", key)
    }
    
    return nil
}

func (p *DatabaseStatePersistence) Exists(key string) bool {
    var exists int
    err := p.existsStmt.QueryRow(key).Scan(&exists)
    return err == nil
}

func (p *DatabaseStatePersistence) StoreBatch(items map[string][]byte) error {
    tx, err := p.db.Begin()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback()
    
    stmt := tx.Stmt(p.insertStmt)
    defer stmt.Close()
    
    for key, data := range items {
        processedData := data
        
        if p.compression {
            compressed, err := p.compress(data)
            if err != nil {
                return fmt.Errorf("compression failed for key %s: %w", key, err)
            }
            processedData = compressed
        }
        
        if p.encryption != nil && p.encryption.Enabled {
            encrypted, err := p.encrypt(processedData)
            if err != nil {
                return fmt.Errorf("encryption failed for key %s: %w", key, err)
            }
            processedData = encrypted
        }
        
        _, err := stmt.Exec(key, processedData)
        if err != nil {
            return fmt.Errorf("failed to store key %s: %w", key, err)
        }
    }
    
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }
    
    // Record metrics
    for key, data := range items {
        p.metrics.RecordWrite(key, len(data))
    }
    
    return nil
}

func (p *DatabaseStatePersistence) LoadBatch(keys []string) (map[string][]byte, error) {
    if len(keys) == 0 {
        return make(map[string][]byte), nil
    }
    
    // Build IN clause
    placeholders := strings.Repeat("?,", len(keys))
    placeholders = placeholders[:len(placeholders)-1] // Remove trailing comma
    
    query := fmt.Sprintf(`
        SELECT %s, %s FROM %s WHERE %s IN (%s)
    `, p.keyColumn, p.dataColumn, p.tableName, p.keyColumn, placeholders)
    
    // Convert keys to interface{} slice
    args := make([]interface{}, len(keys))
    for i, key := range keys {
        args[i] = key
    }
    
    rows, err := p.db.Query(query, args...)
    if err != nil {
        return nil, fmt.Errorf("failed to execute batch query: %w", err)
    }
    defer rows.Close()
    
    result := make(map[string][]byte)
    
    for rows.Next() {
        var key string
        var data []byte
        
        if err := rows.Scan(&key, &data); err != nil {
            return nil, fmt.Errorf("failed to scan row: %w", err)
        }
        
        // Process data
        processedData := data
        
        if p.encryption != nil && p.encryption.Enabled {
            decrypted, err := p.decrypt(processedData)
            if err != nil {
                return nil, fmt.Errorf("decryption failed for key %s: %w", key, err)
            }
            processedData = decrypted
        }
        
        if p.compression {
            decompressed, err := p.decompress(processedData)
            if err != nil {
                return nil, fmt.Errorf("decompression failed for key %s: %w", key, err)
            }
            processedData = decompressed
        }
        
        result[key] = processedData
        
        // Record metrics
        p.metrics.RecordRead(key, len(processedData))
    }
    
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("error during row iteration: %w", err)
    }
    
    return result, nil
}

// Implement compression and encryption methods similar to FileStatePersistence
func (p *DatabaseStatePersistence) compress(data []byte) ([]byte, error) {
    var buf bytes.Buffer
    writer := gzip.NewWriter(&buf)
    
    if _, err := writer.Write(data); err != nil {
        return nil, err
    }
    
    if err := writer.Close(); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}

func (p *DatabaseStatePersistence) decompress(data []byte) ([]byte, error) {
    reader, err := gzip.NewReader(bytes.NewReader(data))
    if err != nil {
        return nil, err
    }
    defer reader.Close()
    
    return ioutil.ReadAll(reader)
}

func (p *DatabaseStatePersistence) encrypt(data []byte) ([]byte, error) {
    // Similar to FileStatePersistence implementation
    if p.encryption == nil || !p.encryption.Enabled {
        return data, nil
    }
    
    block, err := aes.NewCipher(p.encryption.Key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
        return nil, err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, data, nil)
    return ciphertext, nil
}

func (p *DatabaseStatePersistence) decrypt(data []byte) ([]byte, error) {
    // Similar to FileStatePersistence implementation
    if p.encryption == nil || !p.encryption.Enabled {
        return data, nil
    }
    
    block, err := aes.NewCipher(p.encryption.Key)
    if err != nil {
        return nil, err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return nil, err
    }
    
    nonceSize := gcm.NonceSize()
    if len(data) < nonceSize {
        return nil, fmt.Errorf("ciphertext too short")
    }
    
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return nil, err
    }
    
    return plaintext, nil
}
```

---

## State Transactions and ACID Properties

### Transaction Management

```go
// StateTransaction provides ACID transaction support
type StateTransaction interface {
    // Transaction operations
    Get(key string) (interface{}, bool)
    Set(key string, value interface{}) error
    Delete(key string) error
    
    // Transaction control
    Commit() error
    Rollback() error
    IsActive() bool
    
    // Isolation levels
    SetIsolationLevel(level IsolationLevel) error
    GetIsolationLevel() IsolationLevel
    
    // Lock management
    Lock(keys ...string) error
    TryLock(keys ...string) (bool, error)
    
    // Savepoints
    Savepoint(name string) error
    RollbackToSavepoint(name string) error
    ReleaseSavepoint(name string) error
    
    // Metadata
    GetID() string
    GetStartTime() time.Time
    GetState() TransactionState
}

type IsolationLevel int

const (
    ReadUncommitted IsolationLevel = iota
    ReadCommitted
    RepeatableRead
    Serializable
)

type TransactionState string

const (
    TransactionActive    TransactionState = "active"
    TransactionCommitted TransactionState = "committed"
    TransactionAborted   TransactionState = "aborted"
)

// StateTransactionManager manages concurrent transactions
type StateTransactionManager struct {
    state       AgentState
    persistence StatePersistence
    
    // Active transactions
    transactions map[string]*Transaction
    
    // Lock management
    lockManager TransactionLockManager
    
    // Logging
    transactionLog TransactionLog
    
    // Configuration
    maxTransactions int
    defaultTimeout  time.Duration
    
    // Synchronization
    mu     sync.RWMutex
    logger *zap.Logger
}

type Transaction struct {
    id           string
    state        TransactionState
    isolationLevel IsolationLevel
    startTime    time.Time
    timeout      time.Duration
    
    // Transaction data
    readSet      map[string]interface{} // Keys read during transaction
    writeSet     map[string]interface{} // Keys written during transaction
    deleteSet    map[string]bool        // Keys deleted during transaction
    
    // Original values for rollback
    originalValues map[string]interface{}
    
    // Locks held
    locks        []string
    
    // Savepoints
    savepoints   map[string]TransactionSavepoint
    
    // Context
    ctx          context.Context
    cancel       context.CancelFunc
    
    // Manager reference
    manager      *StateTransactionManager
    
    // Synchronization
    mu           sync.RWMutex
}

type TransactionSavepoint struct {
    Name         string                 `json:"name"`
    ReadSet      map[string]interface{} `json:"read_set"`
    WriteSet     map[string]interface{} `json:"write_set"`
    DeleteSet    map[string]bool        `json:"delete_set"`
    CreatedAt    time.Time              `json:"created_at"`
}

func NewStateTransactionManager(state AgentState, persistence StatePersistence) *StateTransactionManager {
    return &StateTransactionManager{
        state:          state,
        persistence:    persistence,
        transactions:   make(map[string]*Transaction),
        lockManager:    NewTransactionLockManager(),
        transactionLog: NewTransactionLog(),
        maxTransactions: 100,
        defaultTimeout:  30 * time.Second,
        logger:         zap.NewNop(),
    }
}

func (tm *StateTransactionManager) BeginTransaction() (*Transaction, error) {
    tm.mu.Lock()
    defer tm.mu.Unlock()
    
    if len(tm.transactions) >= tm.maxTransactions {
        return nil, fmt.Errorf("maximum concurrent transactions reached")
    }
    
    ctx, cancel := context.WithTimeout(context.Background(), tm.defaultTimeout)
    
    tx := &Transaction{
        id:             generateTransactionID(),
        state:          TransactionActive,
        isolationLevel: ReadCommitted,
        startTime:      time.Now(),
        timeout:        tm.defaultTimeout,
        readSet:        make(map[string]interface{}),
        writeSet:       make(map[string]interface{}),
        deleteSet:      make(map[string]bool),
        originalValues: make(map[string]interface{}),
        locks:          make([]string, 0),
        savepoints:     make(map[string]TransactionSavepoint),
        ctx:            ctx,
        cancel:         cancel,
        manager:        tm,
    }
    
    tm.transactions[tx.id] = tx
    
    // Log transaction start
    tm.transactionLog.LogBegin(tx.id, tx.startTime)
    
    tm.logger.Info("Transaction started",
        zap.String("transaction_id", tx.id),
        zap.String("isolation_level", isolationLevelString(tx.isolationLevel)),
    )
    
    return tx, nil
}

func (tx *Transaction) Get(key string) (interface{}, bool) {
    tx.mu.RLock()
    defer tx.mu.RUnlock()
    
    if tx.state != TransactionActive {
        return nil, false
    }
    
    // Check write set first (read your own writes)
    if value, exists := tx.writeSet[key]; exists {
        return value, true
    }
    
    // Check delete set
    if tx.deleteSet[key] {
        return nil, false
    }
    
    // Apply isolation level logic
    switch tx.isolationLevel {
    case ReadUncommitted:
        // Read uncommitted data (dirty reads allowed)
        value, exists := tx.manager.state.Get(key)
        tx.readSet[key] = value
        return value, exists
        
    case ReadCommitted:
        // Read only committed data
        value, exists := tx.manager.state.Get(key)
        tx.readSet[key] = value
        return value, exists
        
    case RepeatableRead:
        // Consistent reads within transaction
        if value, exists := tx.readSet[key]; exists {
            return value, true
        }
        
        value, exists := tx.manager.state.Get(key)
        if exists {
            tx.readSet[key] = value
        }
        return value, exists
        
    case Serializable:
        // Acquire read lock
        if err := tx.Lock(key); err != nil {
            return nil, false
        }
        
        value, exists := tx.manager.state.Get(key)
        if exists {
            tx.readSet[key] = value
        }
        return value, exists
    }
    
    return nil, false
}

func (tx *Transaction) Set(key string, value interface{}) error {
    tx.mu.Lock()
    defer tx.mu.Unlock()
    
    if tx.state != TransactionActive {
        return fmt.Errorf("transaction not active")
    }
    
    // Acquire write lock if needed
    if tx.isolationLevel >= RepeatableRead {
        if err := tx.lockKey(key); err != nil {
            return fmt.Errorf("failed to acquire lock: %w", err)
        }
    }
    
    // Store original value for rollback (if not already stored)
    if _, exists := tx.originalValues[key]; !exists {
        if originalValue, hasOriginal := tx.manager.state.Get(key); hasOriginal {
            tx.originalValues[key] = originalValue
        }
    }
    
    // Add to write set
    tx.writeSet[key] = value
    
    // Remove from delete set if present
    delete(tx.deleteSet, key)
    
    // Log operation
    tx.manager.transactionLog.LogWrite(tx.id, key, value)
    
    return nil
}

func (tx *Transaction) Delete(key string) error {
    tx.mu.Lock()
    defer tx.mu.Unlock()
    
    if tx.state != TransactionActive {
        return fmt.Errorf("transaction not active")
    }
    
    // Acquire write lock if needed
    if tx.isolationLevel >= RepeatableRead {
        if err := tx.lockKey(key); err != nil {
            return fmt.Errorf("failed to acquire lock: %w", err)
        }
    }
    
    // Store original value for rollback (if not already stored)
    if _, exists := tx.originalValues[key]; !exists {
        if originalValue, hasOriginal := tx.manager.state.Get(key); hasOriginal {
            tx.originalValues[key] = originalValue
        }
    }
    
    // Add to delete set
    tx.deleteSet[key] = true
    
    // Remove from write set if present
    delete(tx.writeSet, key)
    
    // Log operation
    tx.manager.transactionLog.LogDelete(tx.id, key)
    
    return nil
}

func (tx *Transaction) Commit() error {
    tx.mu.Lock()
    defer tx.mu.Unlock()
    
    if tx.state != TransactionActive {
        return fmt.Errorf("transaction not active")
    }
    
    // Validate transaction can be committed
    if err := tx.validateCommit(); err != nil {
        return fmt.Errorf("commit validation failed: %w", err)
    }
    
    // Apply changes to state
    for key, value := range tx.writeSet {
        if err := tx.manager.state.Set(key, value); err != nil {
            // Rollback partial changes
            tx.rollbackChanges()
            return fmt.Errorf("failed to apply write for key %s: %w", key, err)
        }
    }
    
    for key := range tx.deleteSet {
        if err := tx.manager.state.Delete(key); err != nil {
            // Continue with deletes - key might not exist
            tx.manager.logger.Warn("Failed to delete key during commit",
                zap.String("transaction_id", tx.id),
                zap.String("key", key),
                zap.Error(err),
            )
        }
    }
    
    // Mark as committed
    tx.state = TransactionCommitted
    
    // Release locks
    tx.releaseLocks()
    
    // Cancel context
    tx.cancel()
    
    // Log commit
    tx.manager.transactionLog.LogCommit(tx.id, time.Now())
    
    // Remove from active transactions
    tx.manager.mu.Lock()
    delete(tx.manager.transactions, tx.id)
    tx.manager.mu.Unlock()
    
    tx.manager.logger.Info("Transaction committed",
        zap.String("transaction_id", tx.id),
        zap.Int("writes", len(tx.writeSet)),
        zap.Int("deletes", len(tx.deleteSet)),
    )
    
    return nil
}

func (tx *Transaction) Rollback() error {
    tx.mu.Lock()
    defer tx.mu.Unlock()
    
    if tx.state != TransactionActive {
        return fmt.Errorf("transaction not active")
    }
    
    // Mark as aborted
    tx.state = TransactionAborted
    
    // Release locks
    tx.releaseLocks()
    
    // Cancel context
    tx.cancel()
    
    // Log rollback
    tx.manager.transactionLog.LogRollback(tx.id, time.Now())
    
    // Remove from active transactions
    tx.manager.mu.Lock()
    delete(tx.manager.transactions, tx.id)
    tx.manager.mu.Unlock()
    
    tx.manager.logger.Info("Transaction rolled back",
        zap.String("transaction_id", tx.id),
    )
    
    return nil
}

func (tx *Transaction) validateCommit() error {
    // Check for conflicts based on isolation level
    switch tx.isolationLevel {
    case Serializable:
        // Check for serialization conflicts
        return tx.checkSerializationConflicts()
    case RepeatableRead:
        // Check for phantom reads
        return tx.checkPhantomReads()
    default:
        return nil
    }
}

func (tx *Transaction) checkSerializationConflicts() error {
    // Check if any keys in read set have been modified by other transactions
    for key := range tx.readSet {
        currentValue, exists := tx.manager.state.Get(key)
        originalValue := tx.readSet[key]
        
        if exists != (originalValue != nil) || (exists && !deepEqual(currentValue, originalValue)) {
            return fmt.Errorf("serialization conflict detected for key %s", key)
        }
    }
    
    return nil
}

func (tx *Transaction) checkPhantomReads() error {
    // Simplified phantom read detection
    // In a full implementation, this would involve range locks
    return nil
}

func (tx *Transaction) rollbackChanges() {
    // Restore original values
    for key, originalValue := range tx.originalValues {
        tx.manager.state.Set(key, originalValue)
    }
}

func (tx *Transaction) lockKey(key string) error {
    if contains(tx.locks, key) {
        return nil // Already locked
    }
    
    if err := tx.manager.lockManager.Lock(tx.id, key); err != nil {
        return err
    }
    
    tx.locks = append(tx.locks, key)
    return nil
}

func (tx *Transaction) releaseLocks() {
    for _, key := range tx.locks {
        tx.manager.lockManager.Unlock(tx.id, key)
    }
    tx.locks = make([]string, 0)
}

func (tx *Transaction) Savepoint(name string) error {
    tx.mu.Lock()
    defer tx.mu.Unlock()
    
    if tx.state != TransactionActive {
        return fmt.Errorf("transaction not active")
    }
    
    // Create savepoint with current state
    savepoint := TransactionSavepoint{
        Name:      name,
        ReadSet:   copyMap(tx.readSet),
        WriteSet:  copyMap(tx.writeSet),
        DeleteSet: copyBoolMap(tx.deleteSet),
        CreatedAt: time.Now(),
    }
    
    tx.savepoints[name] = savepoint
    
    return nil
}

func (tx *Transaction) RollbackToSavepoint(name string) error {
    tx.mu.Lock()
    defer tx.mu.Unlock()
    
    if tx.state != TransactionActive {
        return fmt.Errorf("transaction not active")
    }
    
    savepoint, exists := tx.savepoints[name]
    if !exists {
        return fmt.Errorf("savepoint %s not found", name)
    }
    
    // Restore state to savepoint
    tx.readSet = copyMap(savepoint.ReadSet)
    tx.writeSet = copyMap(savepoint.WriteSet)
    tx.deleteSet = copyBoolMap(savepoint.DeleteSet)
    
    // Remove newer savepoints
    for spName, sp := range tx.savepoints {
        if sp.CreatedAt.After(savepoint.CreatedAt) {
            delete(tx.savepoints, spName)
        }
    }
    
    return nil
}

// Utility functions
func generateTransactionID() string {
    return fmt.Sprintf("tx_%d_%d", time.Now().Unix(), rand.Int63())
}

func isolationLevelString(level IsolationLevel) string {
    switch level {
    case ReadUncommitted:
        return "READ_UNCOMMITTED"
    case ReadCommitted:
        return "READ_COMMITTED"
    case RepeatableRead:
        return "REPEATABLE_READ"
    case Serializable:
        return "SERIALIZABLE"
    default:
        return "UNKNOWN"
    }
}

func contains(slice []string, item string) bool {
    for _, s := range slice {
        if s == item {
            return true
        }
    }
    return false
}

func deepEqual(a, b interface{}) bool {
    // Simplified deep equality check
    // In production, use reflection or a proper deep equality function
    aJSON, _ := json.Marshal(a)
    bJSON, _ := json.Marshal(b)
    return string(aJSON) == string(bJSON)
}

func copyMap(original map[string]interface{}) map[string]interface{} {
    copy := make(map[string]interface{})
    for k, v := range original {
        copy[k] = v
    }
    return copy
}

func copyBoolMap(original map[string]bool) map[string]bool {
    copy := make(map[string]bool)
    for k, v := range original {
        copy[k] = v
    }
    return copy
}
```

---

## State Synchronization and Replication

### State Synchronization

```go
// StateSynchronizer handles state synchronization between agents
type StateSynchronizer interface {
    // Synchronization operations
    SyncState(sourceAgentID, targetAgentID string) error
    SyncStatePartial(sourceAgentID, targetAgentID string, keys []string) error
    
    // Conflict resolution
    ResolveConflicts(agentID string, conflicts []StateConflict) error
    SetConflictResolver(resolver ConflictResolver)
    
    // Change tracking
    TrackChanges(agentID string) (<-chan StateChange, error)
    GetChanges(agentID string, since time.Time) ([]StateChange, error)
    
    // Distributed sync
    SyncWithPeers(agentID string, peers []string) error
    BroadcastChanges(agentID string, change StateChange) error
}

type StateConflict struct {
    Key           string      `json:"key"`
    LocalValue    interface{} `json:"local_value"`
    RemoteValue   interface{} `json:"remote_value"`
    LocalVersion  int64       `json:"local_version"`
    RemoteVersion int64       `json:"remote_version"`
    ConflictType  ConflictType `json:"conflict_type"`
}

type ConflictType string

const (
    ConflictTypeUpdate   ConflictType = "update"   // Both sides modified
    ConflictTypeDelete   ConflictType = "delete"   // One deleted, one modified
    ConflictTypeCreate   ConflictType = "create"   // Both created different values
)

type StateChange struct {
    AgentID     string                 `json:"agent_id"`
    Key         string                 `json:"key"`
    Operation   ChangeOperation        `json:"operation"`
    Value       interface{}            `json:"value,omitempty"`
    OldValue    interface{}            `json:"old_value,omitempty"`
    Version     int64                  `json:"version"`
    Timestamp   time.Time              `json:"timestamp"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type ChangeOperation string

const (
    OperationSet    ChangeOperation = "set"
    OperationDelete ChangeOperation = "delete"
    OperationClear  ChangeOperation = "clear"
)

// ConflictResolver resolves state conflicts
type ConflictResolver interface {
    Resolve(conflict StateConflict) (interface{}, error)
    GetStrategy() ConflictResolutionStrategy
}

type ConflictResolutionStrategy string

const (
    StrategyLastWriteWins    ConflictResolutionStrategy = "last_write_wins"
    StrategyFirstWriteWins   ConflictResolutionStrategy = "first_write_wins"
    StrategyHighestValueWins ConflictResolutionStrategy = "highest_value_wins"
    StrategyMerge           ConflictResolutionStrategy = "merge"
    StrategyManual          ConflictResolutionStrategy = "manual"
)

// DefaultStateSynchronizer implements StateSynchronizer
type DefaultStateSynchronizer struct {
    stateManager     StateManager
    conflictResolver ConflictResolver
    changeTrackers   map[string]*ChangeTracker
    syncPolicy       SyncPolicy
    
    // Communication
    messageBus    MessageBus
    peerRegistry  PeerRegistry
    
    // Monitoring
    metrics *SyncMetrics
    logger  *zap.Logger
    
    // Synchronization
    mu sync.RWMutex
}

type SyncPolicy struct {
    SyncInterval     time.Duration              `json:"sync_interval"`
    ConflictStrategy ConflictResolutionStrategy `json:"conflict_strategy"`
    MaxRetries       int                        `json:"max_retries"`
    RetryBackoff     time.Duration              `json:"retry_backoff"`
    SyncMode         SyncMode                   `json:"sync_mode"`
}

type SyncMode string

const (
    SyncModeImmediate SyncMode = "immediate" // Sync changes immediately
    SyncModeBatch     SyncMode = "batch"     // Batch changes and sync periodically
    SyncModeManual    SyncMode = "manual"    // Sync only when requested
)

func NewStateSynchronizer(stateManager StateManager, policy SyncPolicy) *DefaultStateSynchronizer {
    return &DefaultStateSynchronizer{
        stateManager:     stateManager,
        conflictResolver: NewLastWriteWinsResolver(),
        changeTrackers:   make(map[string]*ChangeTracker),
        syncPolicy:       policy,
        metrics:          NewSyncMetrics(),
        logger:           zap.NewNop(),
    }
}

func (s *DefaultStateSynchronizer) SyncState(sourceAgentID, targetAgentID string) error {
    s.logger.Info("Starting full state sync",
        zap.String("source", sourceAgentID),
        zap.String("target", targetAgentID),
    )
    
    // Load source state
    sourceState, err := s.stateManager.LoadState(sourceAgentID)
    if err != nil {
        return fmt.Errorf("failed to load source state: %w", err)
    }
    
    // Load target state
    targetState, err := s.stateManager.LoadState(targetAgentID)
    if err != nil {
        return fmt.Errorf("failed to load target state: %w", err)
    }
    
    // Get all data from source
    sourceData := sourceState.GetAll()
    targetData := targetState.GetAll()
    
    // Detect conflicts
    conflicts := s.detectConflicts(sourceData, targetData, sourceState, targetState)
    
    // Resolve conflicts
    if len(conflicts) > 0 {
        if err := s.ResolveConflicts(targetAgentID, conflicts); err != nil {
            return fmt.Errorf("conflict resolution failed: %w", err)
        }
    }
    
    // Apply non-conflicting changes
    var syncErrors []error
    for key, value := range sourceData {
        if _, hasConflict := s.findConflict(conflicts, key); !hasConflict {
            if err := targetState.Set(key, value); err != nil {
                syncErrors = append(syncErrors, fmt.Errorf("failed to sync key %s: %w", key, err))
            }
        }
    }
    
    // Save target state
    if err := s.stateManager.SaveState(targetAgentID, targetState); err != nil {
        return fmt.Errorf("failed to save target state: %w", err)
    }
    
    // Record metrics
    s.metrics.RecordSync(sourceAgentID, targetAgentID, len(sourceData), len(conflicts), len(syncErrors))
    
    s.logger.Info("State sync completed",
        zap.String("source", sourceAgentID),
        zap.String("target", targetAgentID),
        zap.Int("keys_synced", len(sourceData)),
        zap.Int("conflicts", len(conflicts)),
        zap.Int("errors", len(syncErrors)),
    )
    
    if len(syncErrors) > 0 {
        return fmt.Errorf("partial sync failure: %d errors", len(syncErrors))
    }
    
    return nil
}

func (s *DefaultStateSynchronizer) detectConflicts(sourceData, targetData map[string]interface{}, sourceState, targetState AgentState) []StateConflict {
    var conflicts []StateConflict
    
    sourceMetadata := sourceState.GetMetadata()
    targetMetadata := targetState.GetMetadata()
    
    // Check for conflicts in keys present in both states
    for key, sourceValue := range sourceData {
        if targetValue, exists := targetData[key]; exists {
            // Compare values
            if !deepEqual(sourceValue, targetValue) {
                conflict := StateConflict{
                    Key:           key,
                    LocalValue:    targetValue,
                    RemoteValue:   sourceValue,
                    LocalVersion:  targetMetadata.Version,
                    RemoteVersion: sourceMetadata.Version,
                    ConflictType:  ConflictTypeUpdate,
                }
                conflicts = append(conflicts, conflict)
            }
        }
    }
    
    return conflicts
}

func (s *DefaultStateSynchronizer) ResolveConflicts(agentID string, conflicts []StateConflict) error {
    state, err := s.stateManager.LoadState(agentID)
    if err != nil {
        return fmt.Errorf("failed to load state: %w", err)
    }
    
    for _, conflict := range conflicts {
        resolvedValue, err := s.conflictResolver.Resolve(conflict)
        if err != nil {
            s.logger.Error("Failed to resolve conflict",
                zap.String("agent_id", agentID),
                zap.String("key", conflict.Key),
                zap.Error(err),
            )
            continue
        }
        
        if err := state.Set(conflict.Key, resolvedValue); err != nil {
            return fmt.Errorf("failed to apply resolved value for key %s: %w", conflict.Key, err)
        }
        
        s.logger.Info("Conflict resolved",
            zap.String("agent_id", agentID),
            zap.String("key", conflict.Key),
            zap.String("strategy", string(s.conflictResolver.GetStrategy())),
        )
    }
    
    return s.stateManager.SaveState(agentID, state)
}

func (s *DefaultStateSynchronizer) findConflict(conflicts []StateConflict, key string) (StateConflict, bool) {
    for _, conflict := range conflicts {
        if conflict.Key == key {
            return conflict, true
        }
    }
    return StateConflict{}, false
}

// LastWriteWinsResolver implements last-write-wins conflict resolution
type LastWriteWinsResolver struct{}

func NewLastWriteWinsResolver() *LastWriteWinsResolver {
    return &LastWriteWinsResolver{}
}

func (r *LastWriteWinsResolver) Resolve(conflict StateConflict) (interface{}, error) {
    // Choose the value from the higher version
    if conflict.RemoteVersion > conflict.LocalVersion {
        return conflict.RemoteValue, nil
    }
    return conflict.LocalValue, nil
}

func (r *LastWriteWinsResolver) GetStrategy() ConflictResolutionStrategy {
    return StrategyLastWriteWins
}

// MergeResolver attempts to merge conflicting values
type MergeResolver struct {
    mergers map[string]ValueMerger
}

type ValueMerger interface {
    Merge(local, remote interface{}) (interface{}, error)
    CanMerge(local, remote interface{}) bool
}

func NewMergeResolver() *MergeResolver {
    resolver := &MergeResolver{
        mergers: make(map[string]ValueMerger),
    }
    
    // Register default mergers
    resolver.mergers["string"] = &StringMerger{}
    resolver.mergers["slice"] = &SliceMerger{}
    resolver.mergers["map"] = &MapMerger{}
    
    return resolver
}

func (r *MergeResolver) Resolve(conflict StateConflict) (interface{}, error) {
    // Determine type of values
    localType := getValueType(conflict.LocalValue)
    remoteType := getValueType(conflict.RemoteValue)
    
    if localType != remoteType {
        // Types don't match, fall back to last-write-wins
        if conflict.RemoteVersion > conflict.LocalVersion {
            return conflict.RemoteValue, nil
        }
        return conflict.LocalValue, nil
    }
    
    // Try to merge using appropriate merger
    if merger, exists := r.mergers[localType]; exists {
        if merger.CanMerge(conflict.LocalValue, conflict.RemoteValue) {
            return merger.Merge(conflict.LocalValue, conflict.RemoteValue)
        }
    }
    
    // Fall back to last-write-wins
    if conflict.RemoteVersion > conflict.LocalVersion {
        return conflict.RemoteValue, nil
    }
    return conflict.LocalValue, nil
}

func (r *MergeResolver) GetStrategy() ConflictResolutionStrategy {
    return StrategyMerge
}

// StringMerger merges string values
type StringMerger struct{}

func (m *StringMerger) CanMerge(local, remote interface{}) bool {
    _, localOK := local.(string)
    _, remoteOK := remote.(string)
    return localOK && remoteOK
}

func (m *StringMerger) Merge(local, remote interface{}) (interface{}, error) {
    localStr, _ := local.(string)
    remoteStr, _ := remote.(string)
    
    // Simple string concatenation
    // In practice, this might use more sophisticated merging like diff3
    return localStr + "\n" + remoteStr, nil
}

// SliceMerger merges slice values
type SliceMerger struct{}

func (m *SliceMerger) CanMerge(local, remote interface{}) bool {
    _, localOK := local.([]interface{})
    _, remoteOK := remote.([]interface{})
    return localOK && remoteOK
}

func (m *SliceMerger) Merge(local, remote interface{}) (interface{}, error) {
    localSlice, _ := local.([]interface{})
    remoteSlice, _ := remote.([]interface{})
    
    // Merge slices, removing duplicates
    merged := append(localSlice, remoteSlice...)
    return removeDuplicates(merged), nil
}

// MapMerger merges map values
type MapMerger struct{}

func (m *MapMerger) CanMerge(local, remote interface{}) bool {
    _, localOK := local.(map[string]interface{})
    _, remoteOK := remote.(map[string]interface{})
    return localOK && remoteOK
}

func (m *MapMerger) Merge(local, remote interface{}) (interface{}, error) {
    localMap, _ := local.(map[string]interface{})
    remoteMap, _ := remote.(map[string]interface{})
    
    // Merge maps, remote values take precedence
    merged := make(map[string]interface{})
    
    // Copy local values
    for k, v := range localMap {
        merged[k] = v
    }
    
    // Override with remote values
    for k, v := range remoteMap {
        merged[k] = v
    }
    
    return merged, nil
}

// Utility functions
func getValueType(value interface{}) string {
    switch value.(type) {
    case string:
        return "string"
    case []interface{}:
        return "slice"
    case map[string]interface{}:
        return "map"
    default:
        return "unknown"
    }
}

func removeDuplicates(slice []interface{}) []interface{} {
    seen := make(map[string]bool)
    result := make([]interface{}, 0)
    
    for _, item := range slice {
        key := fmt.Sprintf("%v", item)
        if !seen[key] {
            seen[key] = true
            result = append(result, item)
        }
    }
    
    return result
}
```

---

## Best Practices

### 1. State Design
- Keep state minimal and focused
- Use clear, consistent naming conventions
- Design for immutability when possible
- Implement proper validation
- Plan for state migration

### 2. Persistence
- Choose appropriate storage backends
- Implement proper error handling
- Use compression for large states
- Consider encryption for sensitive data
- Monitor storage performance

### 3. Transactions
- Use appropriate isolation levels
- Keep transactions short
- Handle deadlocks gracefully
- Implement proper timeout handling
- Plan for rollback scenarios

### 4. Synchronization
- Choose appropriate conflict resolution strategies
- Monitor sync performance
- Handle network partitions
- Implement proper retry logic
- Plan for eventual consistency

### 5. Performance
- Use efficient serialization formats
- Implement proper caching
- Monitor state size
- Optimize for common access patterns
- Plan for horizontal scaling

---

## Next Steps

- **[LLM Agents](llm-agents.md)** - AI-powered agents with tool support
- **[Workflow Agents](workflow-agents.md)** - Sequential, parallel, and conditional patterns
- **[Multi-Agent Systems](multi-agent-systems.md)** - Coordination and communication
- **[Agent Overview](overview.md)** - Agent architecture and concepts
- **[Agent API Reference](../../technical/api-reference/agents.md)** - Detailed API documentation