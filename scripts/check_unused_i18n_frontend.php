<?php

/**
 * 检查前端国际化翻译中未使用的键
 * 新思路：逐个key搜索项目文件，检查是否被使用
 */

class I18nChecker {
    private $frontendPath;
    private $i18nPath;
    private $srcPath;
    private $allKeys = [];
    private $unusedKeys = [];
    
    public function __construct($basePath = '.') {
        $this->frontendPath = rtrim($basePath, '/') . '/frontend';
        $this->i18nPath = $this->frontendPath . '/src/i18n/locales/en-US';
        $this->srcPath = $this->frontendPath . '/src';
    }
    
    /**
     * 主执行方法
     */
    public function run() {
        echo "🔍 检查前端国际化翻译中未使用的键...\n\n";
        
        // 检查路径是否存在
        if (!is_dir($this->i18nPath)) {
            die("❌ 错误: 翻译文件目录不存在: {$this->i18nPath}\n");
        }
        
        if (!is_dir($this->srcPath)) {
            die("❌ 错误: 前端源码目录不存在: {$this->srcPath}\n");
        }
        
        // 1. 读取所有翻译键
        $this->loadAllTranslationKeys();
        
        // 2. 逐个检查键是否被使用
        $this->checkKeyUsage();
        
        // 3. 输出结果
        $this->outputResults();
    }
    
    /**
     * 加载所有翻译键
     */
    private function loadAllTranslationKeys() {
        echo "📖 读取翻译文件...\n";
        
        // 文件名到命名空间的映射
        $namespaceMapping = [
            'adminLogs' => 'adminLogs',
            'api' => 'api',
            'auth' => 'auth',
            'common' => 'common',
            'dashboard' => 'dashboard',
            'devEnvironments' => 'devEnvironments',
            'errors' => 'errors',
            'gitCredentials' => 'gitCredentials',
            'gitDiff' => 'gitDiff',
            'navigation' => 'navigation',
            'projects' => 'projects',
            'systemConfig' => 'systemConfig',
            'taskConversations' => 'taskConversations',
            'tasks' => 'tasks'
        ];
        
        $files = glob($this->i18nPath . '/*.json');
        
        foreach ($files as $file) {
            $fileName = basename($file, '.json');
            $namespace = $namespaceMapping[$fileName] ?? $fileName;
            echo "   - {$fileName}.json (命名空间: {$namespace})\n";
            
            $content = file_get_contents($file);
            $data = json_decode($content, true);
            
            if ($data === null) {
                echo "   ⚠️  警告: 无法解析JSON文件 {$file}\n";
                continue;
            }
            
            $this->extractKeys($data, $namespace, $fileName);
        }
        
        echo "✅ 总共找到 " . count($this->allKeys) . " 个翻译键\n\n";
    }
    
    /**
     * 递归提取所有翻译键
     */
    private function extractKeys($data, $prefix = '', $fileName = '') {
        foreach ($data as $key => $value) {
            $fullKey = $prefix ? $prefix . '.' . $key : $key;
            
            if (is_array($value)) {
                $this->extractKeys($value, $fullKey, $fileName);
            } else {
                // 存储键信息，包括所属文件
                $this->allKeys[$fullKey] = [
                    'fileName' => $fileName,
                    'value' => $value
                ];
            }
        }
    }
    
    /**
     * 逐个检查键的使用情况
     */
    private function checkKeyUsage() {
        echo "🔍 逐个检查翻译键的使用情况...\n";
        
        $totalKeys = count($this->allKeys);
        $checkedKeys = 0;
        
        foreach ($this->allKeys as $key => $keyInfo) {
            $checkedKeys++;
            
            // 显示进度
            if ($checkedKeys % 50 == 0 || $checkedKeys == $totalKeys) {
                $percentage = round(($checkedKeys / $totalKeys) * 100, 1);
                echo "   进度: {$checkedKeys}/{$totalKeys} ({$percentage}%)\n";
            }
            
            // 检查键是否在源码中被使用
            if (!$this->isKeyUsedInSource($key)) {
                $this->unusedKeys[$key] = $keyInfo;
            }
        }
        
        echo "✅ 检查完成！\n\n";
    }
    
    /**
     * 检查单个键是否在源码中被使用
     */
    private function isKeyUsedInSource($key) {
        // 使用 grep 命令在源码目录中搜索该键
        // 转义特殊字符以避免 grep 正则表达式问题
        $escapedKey = escapeshellarg($key);
        $searchPath = escapeshellarg($this->srcPath);
        
        // 搜索包含该键的文件，忽略大小写，递归搜索
        $command = "grep -r -i --include='*.ts' --include='*.tsx' --include='*.js' --include='*.jsx' {$escapedKey} {$searchPath} 2>/dev/null";
        
        // 执行命令并检查是否有输出
        $output = shell_exec($command);
        
        return !empty(trim($output));
    }
    
    /**
     * 输出检查结果
     */
    private function outputResults() {
        echo "🔎 分析结果...\n\n";
        
        if (empty($this->unusedKeys)) {
            echo "🎉 太棒了！所有翻译键都被使用了！\n";
            return;
        }
        
        // 按文件分组显示未使用的键
        $groupedUnused = [];
        foreach ($this->unusedKeys as $key => $keyInfo) {
            $fileName = $keyInfo['fileName'];
            
            if (!isset($groupedUnused[$fileName])) {
                $groupedUnused[$fileName] = [];
            }
            $groupedUnused[$fileName][] = $key; // 保持完整的key名
        }
        
        echo "❌ 发现 " . count($this->unusedKeys) . " 个未使用的翻译键：\n\n";
        
        foreach ($groupedUnused as $fileName => $keys) {
            echo "📁 {$fileName}.json (" . count($keys) . " 个未使用):\n";
            foreach ($keys as $key) {
                echo "   - {$key}\n";
            }
            echo "\n";
        }
        
        // 统计信息
        $totalKeys = count($this->allKeys);
        $unusedCount = count($this->unusedKeys);
        $usedCount = $totalKeys - $unusedCount;
        
        echo "📊 统计信息:\n";
        echo "   - 总翻译键数: {$totalKeys}\n";
        echo "   - 已使用键数: {$usedCount}\n";
        echo "   - 未使用键数: {$unusedCount}\n";
        echo "   - 使用率: " . round(($usedCount / $totalKeys) * 100, 2) . "%\n\n";
        
        // 生成清理建议
        echo "💡 清理建议:\n";
        echo "   可以考虑删除这些未使用的翻译键以减少文件大小\n";
        echo "   删除前请确认这些键确实不会在动态生成的场景中使用\n\n";
        
        echo "🔧 使用方法说明:\n";
        echo "   此脚本使用 grep 命令搜索源码中的翻译键使用情况\n";
        echo "   搜索范围包括 .ts, .tsx, .js, .jsx 文件\n";
        echo "   如果键在任何地方被引用（包括字符串、注释等），都会被认为是已使用\n";
    }
}

// 执行检查
try {
    $checker = new I18nChecker();
    $checker->run();
} catch (Exception $e) {
    echo "❌ 执行出错: " . $e->getMessage() . "\n";
    exit(1);
}

?>