/**
 * MessageQueue.js - Message queue management for MCP Bridge
 * 
 * WHO: MessageQueueManager
 * WHAT: Persistent queue for messages between GitHub and TNOS MCP
 * WHEN: During connection interruptions
 * WHERE: System Layer 2 (Reactive)
 * WHY: To ensure message delivery despite network issues
 * HOW: Using filesystem persistence with compression
 * EXTENT: All cross-system messages
 */

const fs = require('fs');
const path = require('path');
const { MobiusCompression } = require('./MobiusCompression');
const { log } = require('./BridgeLogger');

/**
 * WHO: QueueManager
 * WHAT: MessageQueue class for bidirectional message persistence
 * WHEN: During system operation and connection state changes
 * WHERE: System Layer 2 (Reactive)
 * WHY: To ensure reliable message delivery
 * HOW: Using atomic file operations and compression
 * EXTENT: All MCP messages
 */
class MessageQueue {
  constructor(config) {
    /**
     * WHO: Configuration
     * WHAT: Queue system configuration
     * WHEN: At system initialization
     * WHERE: System Layer 2 (Reactive)
     * WHY: To customize queue behavior
     * HOW: Using configurable parameters
     * EXTENT: Queue system configuration
     */
    this.config = config || {
      queueDir: path.join(__dirname, '..', '..', 'data', 'queues'),
      githubQueueFile: 'github_queue.json',
      tnosQueueFile: 'tnos_queue.json',
      maxQueueSize: 1000,
      flushInterval: 5000, // ms
      compressionLevel: 0.85, // 0-1, higher = more compression
      retryStrategy: 'exponential', // 'linear', 'exponential', 'fixed'
    };

    /**
     * WHO: MessageStorage
     * WHAT: In-memory queue storage
     * WHEN: During system operation
     * WHERE: System memory
     * WHY: For fast message access
     * HOW: Using array data structures
     * EXTENT: All queued messages
     */
    this.github = [];  // Messages to be sent to GitHub MCP
    this.tnos = [];    // Messages to be sent to TNOS MCP

    /**
     * WHO: QueueMetrics
     * WHAT: Performance tracking
     * WHEN: Throughout queue operations
     * WHERE: System Layer 2 (Reactive)
     * WHY: For monitoring and optimization
     * HOW: Using cumulative counters
     * EXTENT: All queue operations
     */
    this.metrics = {
      enqueued: { github: 0, tnos: 0 },
      dequeued: { github: 0, tnos: 0 },
      dropped: { github: 0, tnos: 0 },
      flushes: { github: 0, tnos: 0 },
      errors: { github: 0, tnos: 0 },
      lastFlushTime: { github: 0, tnos: 0 },
      avgQueueTime: { github: 0, tnos: 0 },
      maxQueueTime: { github: 0, tnos: 0 },
      compressionRatio: { github: 0, tnos: 0 }
    };

    // Create queue directory if it doesn't exist
    if (!fs.existsSync(this.config.queueDir)) {
      try {
        fs.mkdirSync(this.config.queueDir, { recursive: true });
        log('info', `Created queue directory: ${this.config.queueDir}`);
      } catch (error) {
        log('error', `Failed to create queue directory: ${error.message}`);
        // Continue with in-memory queues only
      }
    }

    // Setup periodic flush to disk
    this.flushIntervalId = setInterval(() => {
      this.flushToDisk('github');
      this.flushToDisk('tnos');
    }, this.config.flushInterval);

    log('info', 'MessageQueue initialized');
  }

  /**
   * WHO: QueueWriter
   * WHAT: Add a message to the GitHub MCP queue
   * WHEN: When a message needs to be sent to GitHub MCP
   * WHERE: System Layer 2 (Reactive)
   * WHY: To ensure delivery despite connection issues
   * HOW: Using queue data structure with compression
   * EXTENT: All GitHub-bound messages
   */
  queueForGithub(message) {
    return this._enqueue('github', message);
  }

  /**
   * WHO: QueueWriter
   * WHAT: Add a message to the TNOS MCP queue
   * WHEN: When a message needs to be sent to TNOS MCP
   * WHERE: System Layer 2 (Reactive)
   * WHY: To ensure delivery despite connection issues
   * HOW: Using queue data structure with compression
   * EXTENT: All TNOS-bound messages
   */
  queueForTnos(message) {
    return this._enqueue('tnos', message);
  }

  /**
   * WHO: QueueReader
   * WHAT: Get the next message for GitHub MCP
   * WHEN: When a message is ready to be sent to GitHub MCP
   * WHERE: System Layer 2 (Reactive)
   * WHY: For message retrieval from the queue
   * HOW: Using FIFO queue principle
   * EXTENT: Next GitHub-bound message
   */
  dequeueForGithub() {
    return this._dequeue('github');
  }

  /**
   * WHO: QueueReader
   * WHAT: Get the next message for TNOS MCP
   * WHEN: When a message is ready to be sent to TNOS MCP
   * WHERE: System Layer 2 (Reactive)
   * WHY: For message retrieval from the queue
   * HOW: Using FIFO queue principle
   * EXTENT: Next TNOS-bound message
   */
  dequeueForTnos() {
    return this._dequeue('tnos');
  }

  /**
   * WHO: QueueDrain
   * WHAT: Process all queued messages for GitHub MCP
   * WHEN: When connection to GitHub MCP is restored
   * WHERE: System Layer 2 (Reactive)
   * WHY: To clear the queue after reconnection
   * HOW: Using callback processing for each message
   * EXTENT: All queued GitHub-bound messages
   */
  drainGithubQueue(processFunction) {
    return this._drainQueue('github', processFunction);
  }

  /**
   * WHO: QueueDrain
   * WHAT: Process all queued messages for TNOS MCP
   * WHEN: When connection to TNOS MCP is restored
   * WHERE: System Layer 2 (Reactive)
   * WHY: To clear the queue after reconnection
   * HOW: Using callback processing for each message
   * EXTENT: All queued TNOS-bound messages
   */
  drainTnosQueue(processFunction) {
    return this._drainQueue('tnos', processFunction);
  }

  /**
   * WHO: QueuePersistence
   * WHAT: Save all queues to disk
   * WHEN: During shutdown or periodic persistence
   * WHERE: System Layer 2 (Reactive)
   * WHY: To prevent message loss during restarts
   * HOW: Using atomic file writes with compression
   * EXTENT: All queued messages
   */
  async saveQueues() {
    try {
      await Promise.all([
        this.flushToDisk('github'),
        this.flushToDisk('tnos')
      ]);
      log('info', 'All message queues saved to disk');
      return true;
    } catch (error) {
      log('error', `Failed to save queues: ${error.message}`);
      return false;
    }
  }

  /**
   * WHO: QueueRestorer
   * WHAT: Load all queues from disk
   * WHEN: During system startup
   * WHERE: System Layer 2 (Reactive)
   * WHY: To recover messages after system restart
   * HOW: Using file reads with decompression
   * EXTENT: All persisted messages
   */
  loadQueues() {
    try {
      this._loadQueueFromDisk('github');
      this._loadQueueFromDisk('tnos');
      log('info', 'Message queues loaded from disk');
      return true;
    } catch (error) {
      log('error', `Failed to load queues: ${error.message}`);
      return false;
    }
  }

  /**
   * WHO: QueueMonitor
   * WHAT: Get queue statistics
   * WHEN: During monitoring and diagnostics
   * WHERE: System Layer 2 (Reactive)
   * WHY: For system health monitoring
   * HOW: Using accumulated metrics
   * EXTENT: All queue operations
   */
  getMetrics() {
    return {
      ...this.metrics,
      currentSize: {
        github: this.github.length,
        tnos: this.tnos.length
      },
      timestamp: Date.now()
    };
  }

  /**
   * WHO: QueueTerminator
   * WHAT: Clean up resources
   * WHEN: During system shutdown
   * WHERE: System Layer 2 (Reactive)
   * WHY: For proper resource cleanup
   * HOW: Using final flush and interval clearing
   * EXTENT: All queue resources
   */
  async shutdown() {
    clearInterval(this.flushIntervalId);
    await this.saveQueues();
    log('info', 'MessageQueue shutdown complete');
  }

  // Private methods

  /**
   * WHO: InternalQueueWriter
   * WHAT: Internal method to enqueue a message
   * WHEN: During queue operations
   * WHERE: System Layer 2 (Reactive)
   * WHY: For code reuse across queue types
   * HOW: Using common queue logic with timestamps
   * EXTENT: All enqueue operations
   */
  _enqueue(type, message) {
    if (!message) {
      log('warn', `Attempted to queue undefined message for ${type}`);
      return false;
    }

    if (this[type].length >= this.config.maxQueueSize) {
      // Queue is full, apply overflow strategy
      if (this.config.overflowStrategy === 'drop-oldest') {
        this[type].shift(); // Remove oldest message
        this.metrics.dropped[type]++;
        log('warn', `Dropped oldest message from ${type} queue due to overflow`);
      } else if (this.config.overflowStrategy === 'drop-new') {
        this.metrics.dropped[type]++;
        log('warn', `Dropped new message for ${type} queue due to overflow`);
        return false;
      } else {
        // Default: drop-oldest
        this[type].shift();
        this.metrics.dropped[type]++;
        log('warn', `Dropped oldest message from ${type} queue due to overflow`);
      }
    }

    // Add timestamp and context for queue metrics
    const queuedMessage = {
      message,
      queueTime: Date.now(),
      retryCount: 0,
      context: {
        who: 'MessageQueue',
        what: 'Enqueue',
        when: new Date().toISOString(),
        where: `Bridge_${type}Queue`,
        why: 'Connection_Unavailable',
        how: 'Persistence',
        extent: 'Single_Message'
      }
    };

    this[type].push(queuedMessage);
    this.metrics.enqueued[type]++;

    log('debug', `Message queued for ${type}, queue size: ${this[type].length}`);
    return true;
  }

  /**
   * WHO: InternalQueueReader
   * WHAT: Internal method to dequeue a message
   * WHEN: During queue operations
   * WHERE: System Layer 2 (Reactive)
   * WHY: For code reuse across queue types
   * HOW: Using common queue logic with metrics
   * EXTENT: All dequeue operations
   */
  _dequeue(type) {
    if (this[type].length === 0) {
      return null;
    }

    const queuedMessage = this[type].shift();
    this.metrics.dequeued[type]++;

    // Calculate queue time metrics
    const queueTime = Date.now() - queuedMessage.queueTime;

    // Update rolling average queue time
    const previousAvg = this.metrics.avgQueueTime[type];
    const previousCount = this.metrics.dequeued[type] - 1;
    this.metrics.avgQueueTime[type] = previousCount === 0
      ? queueTime
      : (previousAvg * previousCount + queueTime) / this.metrics.dequeued[type];

    // Update max queue time
    if (queueTime > this.metrics.maxQueueTime[type]) {
      this.metrics.maxQueueTime[type] = queueTime;
    }

    log('debug', `Message dequeued from ${type}, queue size: ${this[type].length}`);
    return queuedMessage.message;
  }

  /**
   * WHO: QueueProcessor
   * WHAT: Internal method to process all messages in a queue
   * WHEN: During queue drain operations
   * WHERE: System Layer 2 (Reactive)
   * WHY: For bulk message processing
   * HOW: Using iterative processing with callback
   * EXTENT: All messages in specified queue
   */
  async _drainQueue(type, processFunction) {
    if (!processFunction || typeof processFunction !== 'function') {
      log('error', `Invalid process function provided for ${type} queue drain`);
      return { processed: 0, failed: 0 };
    }

    const metrics = { processed: 0, failed: 0 };

    // Process all messages currently in queue
    const messages = [...this[type]]; // Create a copy to avoid modification issues
    this[type] = []; // Clear the queue

    for (const queuedMessage of messages) {
      try {
        await processFunction(queuedMessage.message);
        metrics.processed++;
      } catch (error) {
        log('error', `Error processing queued message for ${type}: ${error.message}`);

        // Requeue the message with increased retry count
        queuedMessage.retryCount++;

        // Apply retry strategy
        const maxRetries = this.config.maxRetries || 3;
        if (queuedMessage.retryCount <= maxRetries) {
          this[type].push(queuedMessage);
          log('debug', `Requeued message for ${type} (retry ${queuedMessage.retryCount}/${maxRetries})`);
        } else {
          log('warn', `Message for ${type} exceeded max retries, dropped`);
          this.metrics.dropped[type]++;
        }

        metrics.failed++;
      }
    }

    log('info', `Drained ${metrics.processed} messages from ${type} queue (${metrics.failed} failed)`);
    return metrics;
  }

  /**
   * WHO: QueuePersister
   * WHAT: Flush queue to disk
   * WHEN: During persistence operations
   * WHERE: System Layer 2 (Reactive)
   * WHY: For data durability
   * HOW: Using atomic file operations with compression
   * EXTENT: All messages in specified queue
   */
  async flushToDisk(type) {
    if (this[type].length === 0) {
      return true; // Nothing to flush
    }

    try {
      const queueFile = type === 'github' ? this.config.githubQueueFile : this.config.tnosQueueFile;
      const queuePath = path.join(this.config.queueDir, queueFile);
      const tempPath = `${queuePath}.tmp`;

      // Create queue context for compression
      const queueContext = {
        who: 'MessageQueue',
        what: 'Persistence',
        when: new Date().toISOString(),
        where: `Filesystem_${type}Queue`,
        why: 'Data_Durability',
        how: 'Compression',
        extent: 'Full_Queue'
      };

      // Serialize the queue with compression
      const serializedData = JSON.stringify(this[type]);
      const compressedData = MobiusCompression.compress(serializedData, queueContext);

      // Calculate compression ratio for metrics
      const originalSize = Buffer.byteLength(serializedData, 'utf8');
      const compressedSize = Buffer.byteLength(compressedData, 'utf8');
      this.metrics.compressionRatio[type] = originalSize > 0 ? (1 - compressedSize / originalSize) : 0;

      // Write to temporary file first for atomic operation
      await fs.promises.writeFile(tempPath, compressedData);

      // Rename the temporary file to the actual file (atomic operation)
      await fs.promises.rename(tempPath, queuePath);

      this.metrics.flushes[type]++;
      this.metrics.lastFlushTime[type] = Date.now();

      log('debug', `Flushed ${this[type].length} messages for ${type} to disk, compression ratio: ${(this.metrics.compressionRatio[type] * 100).toFixed(2)}%`);
      return true;
    } catch (error) {
      this.metrics.errors[type]++;
      log('error', `Failed to flush ${type} queue to disk: ${error.message}`);
      return false;
    }
  }

  /**
   * WHO: QueueLoader
   * WHAT: Load queue from disk
   * WHEN: During system initialization
   * WHERE: System Layer 2 (Reactive)
   * WHY: For data recovery
   * HOW: Using file reads with decompression
   * EXTENT: All persisted messages
   */
  _loadQueueFromDisk(type) {
    const queueFile = type === 'github' ? this.config.githubQueueFile : this.config.tnosQueueFile;
    const queuePath = path.join(this.config.queueDir, queueFile);

    if (!fs.existsSync(queuePath)) {
      log('info', `No queue file found for ${type}, starting with empty queue`);
      return;
    }

    try {
      // Read and decompress the queue file
      const compressedData = fs.readFileSync(queuePath, 'utf8');

      // Create decompression context
      const queueContext = {
        who: 'MessageQueue',
        what: 'Recovery',
        when: new Date().toISOString(),
        where: `Filesystem_${type}Queue`,
        why: 'Data_Recovery',
        how: 'Decompression',
        extent: 'Full_Queue'
      };

      const decompressedData = MobiusCompression.decompress(compressedData, queueContext);
      const loadedQueue = JSON.parse(decompressedData);

      if (Array.isArray(loadedQueue)) {
        this[type] = loadedQueue;
        log('info', `Loaded ${loadedQueue.length} messages for ${type} queue from disk`);
      } else {
        throw new Error(`Invalid queue data format for ${type}`);
      }
    } catch (error) {
      this.metrics.errors[type]++;
      log('error', `Failed to load ${type} queue from disk: ${error.message}`);
      // Keep the current in-memory queue (empty or previously loaded)
    }
  }
}

module.exports = { MessageQueue };