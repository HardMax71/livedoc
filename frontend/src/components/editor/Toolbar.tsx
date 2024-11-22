import { Component } from 'solid-js';
import Button from '../common/Button';
import { documentStore } from '../../stores/document';

interface ToolbarProps {
    documentId: string;
}

const Toolbar: Component<ToolbarProps> = (props) => {
    const formatDoc = (command: string, value: string | undefined = undefined) => {
        document.execCommand(command, false, value);
    };

    const handleSave = async () => {
        const doc = documentStore.state.currentDocument;
        if (doc) {
            const content = document.querySelector('.editor-content')?.innerHTML || '';
            await documentStore.updateDocument(props.documentId, doc.title, content, doc.version);
        }
    };

    return (
        <div class="editor-toolbar">
            <div class="buttons has-addons">
                <Button onClick={() => formatDoc('bold')} title="Bold">
                    <i class="mdi mdi-format-bold"></i>
                </Button>
                <Button onClick={() => formatDoc('italic')} title="Italic">
                    <i class="mdi mdi-format-italic"></i>
                </Button>
                <Button onClick={() => formatDoc('underline')} title="Underline">
                    <i class="mdi mdi-format-underline"></i>
                </Button>
            </div>

            <div class="buttons has-addons">
                <Button onClick={() => formatDoc('justifyLeft')} title="Align Left">
                    <i class="mdi mdi-format-align-left"></i>
                </Button>
                <Button onClick={() => formatDoc('justifyCenter')} title="Center">
                    <i class="mdi mdi-format-align-center"></i>
                </Button>
                <Button onClick={() => formatDoc('justifyRight')} title="Align Right">
                    <i class="mdi mdi-format-align-right"></i>
                </Button>
            </div>

            <div class="buttons has-addons">
                <Button onClick={() => formatDoc('insertUnorderedList')} title="Bullet List">
                    <i class="mdi mdi-format-list-bulleted"></i>
                </Button>
                <Button onClick={() => formatDoc('insertOrderedList')} title="Numbered List">
                    <i class="mdi mdi-format-list-numbered"></i>
                </Button>
            </div>

            <Button
                variant="primary"
                onClick={handleSave}
                isLoading={documentStore.state.isLoading}
            >
                Save
            </Button>
        </div>
    );
};

export default Toolbar;