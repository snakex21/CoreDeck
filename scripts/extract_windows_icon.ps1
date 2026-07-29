param(
    [Parameter(Mandatory = $true)]
    [string]$SourcePath,

    [Parameter(Mandatory = $true)]
    [string]$OutputPath
)

$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Drawing

if (-not (Test-Path -LiteralPath $SourcePath -PathType Leaf)) {
    throw "Source file does not exist: $SourcePath"
}

$nativeCode = @'
using System;
using System.Runtime.InteropServices;

public static class ShellJumboIcon
{
    [StructLayout(LayoutKind.Sequential, CharSet = CharSet.Unicode)]
    private struct SHFILEINFO
    {
        public IntPtr hIcon;
        public int iIcon;
        public uint dwAttributes;
        [MarshalAs(UnmanagedType.ByValTStr, SizeConst = 260)]
        public string szDisplayName;
        [MarshalAs(UnmanagedType.ByValTStr, SizeConst = 80)]
        public string szTypeName;
    }

    [StructLayout(LayoutKind.Sequential)]
    private struct IMAGELISTDRAWPARAMS
    {
        public int cbSize;
        public IntPtr himl;
        public int i;
        public IntPtr hdcDst;
        public int x;
        public int y;
        public int cx;
        public int cy;
        public int xBitmap;
        public int yBitmap;
        public uint rgbBk;
        public uint rgbFg;
        public uint fStyle;
        public uint dwRop;
        public uint fState;
        public uint Frame;
        public uint crEffect;
    }

    [ComImport]
    [Guid("46EB5926-582E-4017-9FDF-E8998DAA0950")]
    [InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
    private interface IImageList
    {
        [PreserveSig] int Add(IntPtr hbmImage, IntPtr hbmMask, ref int pi);
        [PreserveSig] int ReplaceIcon(int i, IntPtr hicon, ref int pi);
        [PreserveSig] int SetOverlayImage(int iImage, int iOverlay);
        [PreserveSig] int Replace(int i, IntPtr hbmImage, IntPtr hbmMask);
        [PreserveSig] int AddMasked(IntPtr hbmImage, int crMask, ref int pi);
        [PreserveSig] int Draw(ref IMAGELISTDRAWPARAMS pimldp);
        [PreserveSig] int Remove(int i);
        [PreserveSig] int GetIcon(int i, int flags, out IntPtr picon);
    }

    [DllImport("shell32.dll", CharSet = CharSet.Unicode)]
    private static extern IntPtr SHGetFileInfo(
        string pszPath,
        uint dwFileAttributes,
        ref SHFILEINFO psfi,
        uint cbFileInfo,
        uint uFlags);

    [DllImport("shell32.dll", EntryPoint = "#727")]
    private static extern int SHGetImageList(
        int iImageList,
        ref Guid riid,
        out IImageList ppv);

    [DllImport("user32.dll", SetLastError = true)]
    public static extern bool DestroyIcon(IntPtr hIcon);

    public static IntPtr GetIconHandle(string path)
    {
        const uint SHGFI_SYSICONINDEX = 0x000004000;
        const int SHIL_JUMBO = 4;
        const int ILD_TRANSPARENT = 1;

        SHFILEINFO info = new SHFILEINFO();
        IntPtr result = SHGetFileInfo(
            path,
            0,
            ref info,
            (uint)Marshal.SizeOf(info),
            SHGFI_SYSICONINDEX);

        if (result == IntPtr.Zero)
            throw new InvalidOperationException("SHGetFileInfo failed.");

        Guid iid = new Guid("46EB5926-582E-4017-9FDF-E8998DAA0950");
        IImageList imageList;
        int hr = SHGetImageList(SHIL_JUMBO, ref iid, out imageList);
        if (hr != 0)
            Marshal.ThrowExceptionForHR(hr);

        IntPtr iconHandle;
        hr = imageList.GetIcon(info.iIcon, ILD_TRANSPARENT, out iconHandle);
        if (hr != 0)
            Marshal.ThrowExceptionForHR(hr);

        return iconHandle;
    }
}
'@

if (-not ('ShellJumboIcon' -as [type])) {
    Add-Type -TypeDefinition $nativeCode
}

$outputDirectory = Split-Path -Parent $OutputPath
if ($outputDirectory -and -not (Test-Path -LiteralPath $outputDirectory)) {
    New-Item -ItemType Directory -Path $outputDirectory | Out-Null
}

$handle = [ShellJumboIcon]::GetIconHandle((Resolve-Path -LiteralPath $SourcePath).Path)
try {
    $icon = [System.Drawing.Icon]::FromHandle($handle)
    $bitmap = $icon.ToBitmap()
    try {
        $bitmap.Save($OutputPath, [System.Drawing.Imaging.ImageFormat]::Png)
    }
    finally {
        $bitmap.Dispose()
        $icon.Dispose()
    }
}
finally {
    [void][ShellJumboIcon]::DestroyIcon($handle)
}

Write-Output $OutputPath
