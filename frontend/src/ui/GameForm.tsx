import { useEffect, useRef, useState } from 'react'
import { currentUser } from '../api/authentication'

const SIZE = 15
type GameState = {
  resultSaved: boolean
  board: number[]
  turn: number
  color: number
  winner: number
  finished: boolean
  reason: string
  lastMove: number
}

function GameForm() {
  const mode = new URLSearchParams(window.location.search).get('mode') === 'online' ? 'online' : 'cpu'
  const socket = useRef<WebSocket | null>(null)
  const finished = useRef(false)
  const [game, setGame] = useState<GameState | null>(null)
  const [status, setStatus] = useState('ログイン状態を確認中…')
  const [error, setError] = useState('')
  const [closed, setClosed] = useState(false)
  const [pending, setPending] = useState(false)
  const [isWaiting, setIsWaiting] = useState(false)
  const [selectedCell, setSelectedCell] = useState<number | null>(null)
  const [cpuColor, setCPUColor] = useState<'black' | 'white' | null>(null)
  const choosingColor = mode === 'cpu' && cpuColor === null

  useEffect(() => {
    if (mode === 'cpu' && cpuColor === null) return
    let cancelled = false
    let connection: WebSocket | null = null
    const abort = new AbortController()
    async function connect() {
      try {
        if (await currentUser(abort.signal) === null) {
          if (!cancelled) { setStatus('ログインが必要です。'); setClosed(true) }
          return
        }
        if (cancelled) return
        setStatus('接続中…')
        const colorQuery = mode === 'cpu' ? `&color=${cpuColor}` : ''
        connection = new WebSocket(`${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/game/ws?mode=${mode}${colorQuery}`)
        socket.current = connection
        connection.onmessage = (event) => {
          if (cancelled) return
          const message = JSON.parse(event.data)
          setPending(false)
          if (message.type === 'waiting') {
            setIsWaiting(true)
            setStatus(message.message || '対戦相手を待っています。')
            setError('')
          }
          if (message.type === 'state') {
            setIsWaiting(false)
            finished.current = message.finished
            setGame(message)
            setError('')
          }
          if (message.type === 'error') setError(message.message)
        }
        connection.onclose = () => {
          if (cancelled) return
          setClosed(true)
          setIsWaiting(false)
          setPending(false)
          if (!finished.current) setError('接続が切れました。再接続するか、ログイン状態・他の対戦タブを確認してください。対局中の切断は負けになります。')
        }
        connection.onerror = () => {
          if (!cancelled) setError('対戦サーバーとの通信に失敗しました。')
        }
      } catch {
        if (!cancelled) { setError('ログイン状態を確認できませんでした。'); setClosed(true) }
      }
    }
    void connect()
    return () => { cancelled = true; abort.abort(); connection?.close(); socket.current = null }
  }, [mode, cpuColor])

  function send(type: 'move' | 'resign', index?: number) {
    if (socket.current?.readyState !== WebSocket.OPEN || pending) return
    setPending(true)
    socket.current.send(JSON.stringify({ type, index }))
  }

  const canMove = !!game && !game.finished && game.turn === game.color && !closed && !pending
  const canConfirm = canMove && selectedCell !== null && game.board[selectedCell] === 0
  function chooseCell(index: number) {
    if (window.matchMedia('(max-width: 640px), (pointer: coarse)').matches) {
      setSelectedCell(index)
    } else {
      send('move', index)
    }
  }
  const message = !game ? closed ? '対戦に接続していません' : status : game.finished
    ? game.winner === 0 ? '引き分けです' : game.winner === game.color ? 'あなたの勝ちです！' : 'あなたの負けです'
    : closed ? '対局との接続が切れました' : game.turn === game.color ? 'あなたの番です' : mode === 'cpu' ? 'CPUの番です' : '相手の番です'
  const board = game?.board ?? Array<number>(SIZE * SIZE).fill(0)
  const moveCount = board.filter((cell) => cell !== 0).length
  const opponent = mode === 'cpu' ? 'CPU' : '対戦相手'
  const result = game?.finished ? game.winner === 0 ? 'draw' : game.winner === game.color ? 'win' : 'lose' : null
  const resultReason = game?.reason === 'disconnect' ? '相手の切断により対局が終了しました。'
    : game?.reason === 'resign' ? `${game.winner === game.color ? opponent : 'あなた'}が投了しました。`
    : game?.reason === 'draw' ? '盤面がすべて埋まりました。'
    : '縦・横・斜めのいずれかに、ちょうど5個並びました。'

  return (
    <main className="game-page">
      <section className="game-screen" aria-labelledby="game-title">
        <header className="game-header">
          <a className="game-back-link" href="/home">ホームへ戻る</a>
          <p className="home-eyebrow">GOMOKU TAISEN</p>
          <h1 id="game-title">五目並べ · {mode === 'cpu' ? 'CPU対戦' : 'オンライン対戦'}</h1>
        </header>
        {choosingColor && (
          <section className="game-waiting" aria-labelledby="cpu-color-title">
            <h2 id="cpu-color-title">先手・後手を選んでください</h2>
            <p>黒が先に打ちます。後手を選ぶとCPUの初手から始まります。</p>
            <div className="game-controls">
              <button type="button" className="home-button home-button-primary" onClick={() => setCPUColor('black')}>先手（黒）で始める</button>
              <button type="button" className="home-button home-button-secondary" onClick={() => setCPUColor('white')}>後手（白）で始める</button>
            </div>
          </section>
        )}
        {!choosingColor && !game && !closed && (
          <div className="game-waiting" role="status" aria-live="polite">
            <span className="game-waiting-dot" aria-hidden="true" />
            <strong>{status}</strong>
            <p>{isWaiting ? '現在、対戦できる相手がいません。この画面のままお待ちください。相手が参加すると自動で対局が始まります。' : '対局の準備ができるまでお待ちください。'}</p>
            {isWaiting && <a className="home-button home-button-secondary" href="/home">待機をキャンセルしてホームへ戻る</a>}
          </div>
        )}
        {(game || closed) && !result && <p className="game-turn" role="status">{pending ? '送信中…' : message}</p>}
        {result && (
          <section className={`game-result game-result-${result}`} role="status" aria-label="対戦結果">
            <p className="game-result-label">対戦結果</p>
            <h2>{message}</h2>
            <p>{resultReason}</p>
            <p>{moveCount}手で終了</p>
            <p>{game?.resultSaved ? '戦績・レートを保存しました。' : '戦績・レートを保存中…'}</p>
          </section>
        )}
        {game && (
          <div className="game-players" aria-label="対戦者と石の色">
            <div className={`game-player${!game.finished && game.turn === game.color ? ' game-player-active' : ''}`}>
              <span className={`game-player-stone game-player-stone-${game.color === 1 ? 'black' : 'white'}`} aria-hidden="true" />
              <span>あなた <small>{game.color === 1 ? '黒・先手' : '白・後手'}</small></span>
            </div>
            <span className="game-move-count">{moveCount}手</span>
            <div className={`game-player${!game.finished && game.turn !== game.color ? ' game-player-active' : ''}`}>
              <span className={`game-player-stone game-player-stone-${game.color === 1 ? 'white' : 'black'}`} aria-hidden="true" />
              <span>{opponent} <small>{game.color === 1 ? '白・後手' : '黒・先手'}</small></span>
            </div>
          </div>
        )}
        {error && <p role="alert" className="home-status">{error}</p>}
        {game && <div className="gomoku-board" role="group" aria-label="五目並べの盤面">
          {board.map((cell, index) => (
            <button key={index} type="button"
              className={`gomoku-cell${cell ? ` gomoku-stone-${cell === 1 ? 'black' : 'white'}` : ''}${game?.lastMove === index ? ' gomoku-last-move' : ''}${canConfirm && selectedCell === index ? ' gomoku-cell-selected' : ''}`}
              style={{ left: `${5 + (index % SIZE / (SIZE - 1)) * 90}%`, top: `${5 + (Math.floor(index / SIZE) / (SIZE - 1)) * 90}%` }}
              aria-label={`${Math.floor(index / SIZE) + 1}行${index % SIZE + 1}列 ${cell === 0 ? '空き' : cell === 1 ? '黒' : '白'}${game?.lastMove === index ? '（最後の着手）' : ''}`}
              disabled={!canMove || cell !== 0} onClick={() => chooseCell(index)}>
              {cell !== 0 && <span className="gomoku-stone" aria-hidden="true" />}
            </button>
          ))}
        </div>}
        {game && !game.finished && (
          <div className="game-touch-controls">
            <p className="game-instructions">交点をタップして選び、「ここに打つ」で確定します。</p>
            <p className="game-instructions" role="status">{canConfirm && selectedCell !== null ? `${Math.floor(selectedCell / SIZE) + 1}行${selectedCell % SIZE + 1}列を選択中` : '打つ場所を選んでください'}</p>
            <button type="button" className="home-button home-button-primary" disabled={!canConfirm} onClick={() => {
              if (canConfirm && selectedCell !== null) { send('move', selectedCell); setSelectedCell(null) }
            }}>ここに打つ</button>
          </div>
        )}
        <div className="game-controls">
          {game && !game.finished && !closed && <button className="home-button home-button-secondary" disabled={pending} onClick={() => send('resign')}>投了する</button>}
          {(game?.finished || closed) && <a className="home-button home-button-primary" href={`/game?mode=${mode}`}>{closed && !game?.finished ? '新しい対局を始める' : mode === 'online' ? '次の相手を探す' : 'もう一度CPUと対戦'}</a>}
          {game?.finished && <a className="home-button home-button-secondary" href="/home">ホームで対戦方式を選ぶ</a>}
          {!game && !closed && !isWaiting && <a className="home-button home-button-secondary" href="/home">ホームへ戻る</a>}
          {closed && <a href="/login">ログイン画面へ</a>}
        </div>
        <details className="game-rules">
          <summary>対局ルール</summary>
          <ul>
            <li>縦・横・斜めにちょうど5個並べると勝ちです。</li>
            <li>6個以上の連続では勝ちにならず、対局を続けます。</li>
            <li>盤面がすべて埋まると引き分けです。</li>
            <li>{mode === 'cpu' ? 'CPU対戦では開始前に先手・後手を選べます。' : 'オンライン対戦の先手は自動で決まります。'}赤い印は最後の着手です。</li>
            <li>投了・対局中の退出・切断は負けになります。</li>
          </ul>
        </details>
      </section>
    </main>
  )
}

export default GameForm
