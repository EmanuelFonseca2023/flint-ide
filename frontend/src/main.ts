import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker'
import * as monaco from 'monaco-editor'
import { EventsOn } from '../wailsjs/runtime/runtime'
import {
  Compile, SendInput, KillProcess,
  CompilerInfo, OpenFile, SaveFile, SaveFileAs, CurrentFileName
} from '../wailsjs/go/main/App'

self.MonacoEnvironment = {
  getWorker(_: any, _label: string) {
    return new EditorWorker()
  }
}

const DEFAULT_CODE = `#include <iostream>
using namespace std;

int main() {
    cout << "Hola desde Flint!" << endl;
    return 0;
}
`

// ── Monaco ────────────────────────────────────────────────────────────────────
const editor = monaco.editor.create(document.getElementById('editor-pane')!, {
  value: DEFAULT_CODE,
  language: 'cpp',
  theme: 'vs-dark',
  fontSize: 14,
  fontFamily: "'Menlo', 'Consolas', monospace",
  lineHeight: 22,
  minimap: { enabled: false },
  scrollBeyondLastLine: false,
  renderLineHighlight: 'line',
  cursorBlinking: 'smooth',
  smoothScrolling: true,
  padding: { top: 12, bottom: 12 },
  automaticLayout: true,
})

// ── Estado ────────────────────────────────────────────────────────────────────
let errorDecorations: string[] = []
let isUnsaved = false

// ── DOM ───────────────────────────────────────────────────────────────────────
const elFileName   = document.getElementById('status-filename')!
const elDot        = document.getElementById('dot-modified')!
const elCompDot    = document.getElementById('compiler-dot')!
const elCompLabel  = document.getElementById('compiler-label')!
const elCursor     = document.getElementById('status-cursor')!
const elStatusErr  = document.getElementById('status-errors')!
const elStatusSep  = document.getElementById('status-sep-err')!
const elStatusLang = document.getElementById('status-lang')!
const elOutput     = document.getElementById('io-output')!
const elErrList    = document.getElementById('err-list')!
const elErrCount   = document.getElementById('err-count')!
const elBtnRun     = document.getElementById('btn-run')!
const elInputRow   = document.getElementById('io-input-row')!
const elInput      = document.getElementById('io-input') as HTMLInputElement

// ── Compilador ────────────────────────────────────────────────────────────────
async function initCompiler() {
  try {
    const info = await CompilerInfo()
    if (info.Found) {
      elCompDot.classList.add('ready')
      const match = info.Version.match(/(gcc|clang|g\+\+)[^\d]*(\d+\.\d+)/i)
      const name  = match ? match[1].toLowerCase() : 'compilador'
      const ver   = match ? match[2] : ''
      elCompLabel.textContent = `${name} ${ver}`.trim()
      elStatusLang.textContent = `C++17 · ${name}`
    } else {
      elCompDot.classList.add('error')
      elCompLabel.textContent = 'Compilador no encontrado'
    }
  } catch {
    elCompDot.classList.add('error')
    elCompLabel.textContent = 'Error al detectar compilador'
  }
}

// ── Cursor ────────────────────────────────────────────────────────────────────
editor.onDidChangeCursorPosition((e) => {
  const { lineNumber, column } = e.position
  elCursor.textContent = `Línea ${lineNumber}, Col ${column}`
})

editor.onDidChangeModelContent(() => {
  if (!isUnsaved) {
    isUnsaved = true
    elDot.classList.add('unsaved')
  }
})

// ── Compilar ──────────────────────────────────────────────────────────────────
function runCode() {
  setButtonDetener()
  elOutput.textContent = 'Compilando...\n'
  elInputRow.style.display = 'none'
  elErrList.innerHTML = ''
  elErrCount.style.display = 'none'
  elStatusErr.style.display = 'none'
  elStatusSep.style.display = 'none'
  clearErrorDecorations()
  Compile(editor.getValue())
}

// ── Eventos del backend ───────────────────────────────────────────────────────
EventsOn('compile:start', () => {
  elOutput.textContent = 'Compilando...\n'
})

EventsOn('compile:ok', () => {
  elOutput.textContent = ''
  elErrList.innerHTML = '<span class="no-errors">Sin errores de compilación</span>'
  elInputRow.style.display = 'flex'
  elInput.focus()
})

EventsOn('compile:error', (errors: any[]) => {
  elOutput.textContent = ''
  showErrors(errors)
  resetRunButton()
})

EventsOn('process:output', (event: { Type: string; Data: string }) => {
  if (event.Type === 'stdout' || event.Type === 'stderr') {
    elOutput.textContent += event.Data
    const body = elOutput.parentElement!
    body.scrollTop = body.scrollHeight
  }
  if (event.Type === 'exit') {
    elOutput.textContent += '\n─────────────────────\n[Proceso terminó]\n'
    elInputRow.style.display = 'none'
    resetRunButton()
  }
})

// ── Input del usuario ─────────────────────────────────────────────────────────
elInput.addEventListener('keydown', async (e) => {
  if (e.key === 'Enter') {
    const val = elInput.value
    elInput.value = ''
    elOutput.textContent += val + '\n'
    const body = elOutput.parentElement!
    body.scrollTop = body.scrollHeight
    await SendInput(val)
  }
})

// ── Botones ───────────────────────────────────────────────────────────────────
function setButtonDetener() {
  elBtnRun.innerHTML = '⏹ Detener'
  elBtnRun.onclick = () => {
    KillProcess()
    elInputRow.style.display = 'none'
    resetRunButton()
  }
}

function resetRunButton() {
  elBtnRun.innerHTML = `<svg width="13" height="13" viewBox="0 0 24 24" fill="currentColor">
    <polygon points="5 3 19 12 5 21 5 3"/>
  </svg> Compilar y Ejecutar`
  elBtnRun.onclick = () => runCode()
}

// ── Errores ───────────────────────────────────────────────────────────────────
function showErrors(errors: any[]) {
  const count = errors.length
  elErrCount.textContent = String(count)
  elErrCount.style.display = 'inline'
  elStatusErr.textContent = `${count} error${count > 1 ? 'es' : ''}`
  elStatusErr.style.display = 'inline'
  elStatusSep.style.display = 'inline'

  elErrList.innerHTML = ''
  errors.forEach((err) => {
    const item = document.createElement('div')
    item.className = 'err-item'
    const lines = err.Message.split('\n')
    const main  = lines[0]
    const hint  = lines.find((l: string) => l.includes('→')) || ''
    item.innerHTML = `
      <span class="err-line-badge">${err.Line > 0 ? `L${err.Line}` : '—'}</span>
      <div>
        <div class="err-msg">${main}</div>
        ${hint ? `<div class="err-hint">${hint}</div>` : ''}
      </div>`
    if (err.Line > 0) {
      item.style.cursor = 'pointer'
      item.addEventListener('click', () => {
        editor.revealLineInCenter(err.Line)
        editor.setPosition({ lineNumber: err.Line, column: err.Column || 1 })
        editor.focus()
      })
    }
    elErrList.appendChild(item)
  })

  highlightErrorLines(errors)
}

function highlightErrorLines(errors: any[]) {
  clearErrorDecorations()
  errorDecorations = editor.deltaDecorations([],
    errors.filter(e => e.Line > 0).map(e => ({
      range: new monaco.Range(e.Line, 1, e.Line, 1),
      options: { isWholeLine: true, className: 'monaco-error-line' }
    }))
  )
}

function clearErrorDecorations() {
  if (errorDecorations.length > 0) {
    editor.deltaDecorations(errorDecorations, [])
    errorDecorations = []
  }
}

// ── Archivos ──────────────────────────────────────────────────────────────────
async function openFile() {
  try {
    const content = await OpenFile()
    if (content !== '') {
      editor.setValue(content)
      elFileName.textContent = await CurrentFileName()
      isUnsaved = false
      elDot.classList.remove('unsaved')
    }
  } catch (e) { console.error(e) }
}

async function saveFile() {
  try {
    await SaveFile(editor.getValue())
    elFileName.textContent = await CurrentFileName()
    isUnsaved = false
    elDot.classList.remove('unsaved')
  } catch (e) { console.error(e) }
}

async function saveFileAs() {
  try {
    await SaveFileAs(editor.getValue())
    elFileName.textContent = await CurrentFileName()
    isUnsaved = false
    elDot.classList.remove('unsaved')
  } catch (e) { console.error(e) }
}

// ── Atajos ────────────────────────────────────────────────────────────────────
document.addEventListener('keydown', (e) => {
  if (e.metaKey || e.ctrlKey) {
    if (e.key === 's') { e.preventDefault(); e.shiftKey ? saveFileAs() : saveFile() }
    if (e.key === 'o') { e.preventDefault(); openFile() }
    if (e.key === 'Enter') { e.preventDefault(); runCode() }
  }
})

document.getElementById('btn-open')!.addEventListener('click', openFile)
document.getElementById('btn-save')!.addEventListener('click', saveFile)
document.getElementById('btn-clear-output')!.addEventListener('click', () => {
  elOutput.textContent = ''
})
elBtnRun.addEventListener('click', runCode)

// ── CSS errores ───────────────────────────────────────────────────────────────
const style = document.createElement('style')
style.textContent = `.monaco-error-line {
  background: rgba(232, 108, 44, 0.12) !important;
  border-left: 2px solid #E86C2C !important;
}`
document.head.appendChild(style)

// ── Resizers ──────────────────────────────────────────────────────────────────
// ── Resizers ──────────────────────────────────────────────────────────────────
function initResizers() {
  const resizerH   = document.getElementById('resizer-h')!
  const rightPane  = document.querySelector('.right-pane') as HTMLElement

  resizerH.addEventListener('mousedown', (e) => {
    e.preventDefault()
    resizerH.classList.add('dragging')
    const startX      = e.clientX
    const startWidth  = rightPane.offsetWidth

    const onMove = (e: MouseEvent) => {
      const delta    = startX - e.clientX
      const newWidth = Math.max(180, Math.min(startWidth + delta, window.innerWidth * 0.6))
      rightPane.style.flex = `0 0 ${newWidth}px`
    }
    const onUp = () => {
      resizerH.classList.remove('dragging')
      document.removeEventListener('mousemove', onMove)
      document.removeEventListener('mouseup', onUp)
    }
    document.addEventListener('mousemove', onMove)
    document.addEventListener('mouseup', onUp)
  })

  window.addEventListener('resize', () => {
    const maxW = window.innerWidth * 0.6
    if (rightPane.offsetWidth > maxW) {
      rightPane.style.flex = `0 0 ${maxW}px`
    }
  })

  const resizerV = document.getElementById('resizer-v')!
  const ioPanel  = document.querySelector('.io-panel') as HTMLElement

  resizerV.addEventListener('mousedown', (e) => {
    e.preventDefault()
    resizerV.classList.add('dragging')
    const startY      = e.clientY
    const startHeight = ioPanel.offsetHeight

    const onMove = (e: MouseEvent) => {
      const newHeight = Math.max(80, startHeight + (e.clientY - startY))
      ioPanel.style.flex   = 'none'
      ioPanel.style.height = newHeight + 'px'
    }
    const onUp = () => {
      resizerV.classList.remove('dragging')
      document.removeEventListener('mousemove', onMove)
      document.removeEventListener('mouseup', onUp)
    }
    document.addEventListener('mousemove', onMove)
    document.addEventListener('mouseup', onUp)
  })
}

// ── Init ──────────────────────────────────────────────────────────────────────
initResizers()
initCompiler()
resetRunButton()